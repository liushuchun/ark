package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/log.v1"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"qbox.us/cc/config"
	"qiniu.ai/lib/model"
	"qiniu.ai/video/models"
	"qiniu.com/auth/qiniumac.v1"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	//AK  a  key
	AK = ""
	//SK   secret key
	SK         = ""
	bucketHost = "p1f56xgi8.bkt.clouddn.com"
	bucket     = "video"
)

var (
	err             error
	msgsChan        = make(chan Job, 50000)
	pwd             string
	workspace       string
	mac             *qbox.Mac
	pubPolicy       storage.PutPolicy
	cfg             storage.Config
	ATLAB_HOST_PURE = "http://serve.atlab.ai"
	ATLAB_HOST      = "http://serve.atlab.ai/v1/"
	SCENE_API       = ATLAB_HOST + "eval/scene"
	FACE_API        = ATLAB_HOST + "eval/facex-detect"
	DETECTION_API   = ATLAB_HOST + "eval/detection"
	BATCH_API       = ATLAB_HOST + "batch"
	OCR_API         = "127.0.0.1:8010"
	VOICE_API       = "127.0.0.1:8009"
	chunkSize       = 5
	logger          *log.Logger

	type2APIMap = map[string]string{
		"scene":  "/v1/eval/scene",
		"voice":  "/v1/eval/voice",
		"people": "http://argus.atlab.ai/v1/celebrity/search",
		"object": "/v1/eval/detection",
	}
)

type (
	Config struct {
		Mgo        model.Config `json:"mgo"`
		BindHost   string       `json:"bind_host"`
		AK         string       `json:"ak"`
		SK         string       `json:"sk"`
		AtlabHost  string       `json:"atlab_host"`
		Bucket     string       `json:"bucket"`
		BktHost    string       `json:"bucket_host"`
		MaxProcs   int          `json:"max_procs"`
		DebugLevel int          `json:"debug_level"`
	}

	videoRequest struct {
		Src      string `json:"src"`
		Name     string `json:"name"`
		Choice   string `json:"choice"`
		CallBack string `json:"callback,omitempty"`
	}

	Job struct {
		fileURI  string
		name     string
		choices  []string
		callBack string
		id       bson.ObjectId
	}

	BatchReqParam struct {
		Op   string `json:"op"`
		Data Data   `json:"data"`
	}
	ReqParam struct {
		Data Data `json:"data"`
	}
	Data struct {
		URI string `json:"uri"`
	}
	//result

	RespBody struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Result  interface{} `json:"result"`
	}

	ApiResult struct {
		Index int      `json:"index"`
		Class string   `json:"class"`
		Score float64  `json:"score"`
		Label []string `json:"label",omitempty`
	}
)

func do(msg Job, workerPath string, conf *Config) (err error) {
	if _, err = os.Stat(workerPath); os.IsNotExist(err) {
		if errWorker := os.Mkdir(workerPath, os.ModePerm); errWorker != nil {
			logger.Errorf("os.Mkdir(%s,os.ModePerm) with error:%v \n", workerPath, errWorker)
			return
		}

	}
	workerImagePath := path.Join(workerPath, "images")
	if _, err = os.Stat(workerImagePath); os.IsNotExist(err) {
		errMkdir := os.Mkdir(workerImagePath, os.ModePerm)
		if errMkdir != nil {
			logger.Errorf("os.Mkdir(%s,os.ModePerm) with error:%v \n", workerImagePath, errMkdir)
			return
		}
	}
	defer os.RemoveAll(workerPath)
	log.Printf("the msg in->%s\n", msg.name)

	fileName, err := download(msg.fileURI, workerPath)
	if err != nil {
		time.Sleep(time.Second * 10)
		fileName, err = download(msg.fileURI, workerPath)
		if err != nil {

			return err
		}
	}
	//cmd := fmt.Sprintf("ffmpeg -i %s -r 1 -f image2 %s/%s-%%5d.jpg", fileName, workerImagePath, msg.id.Hex())

	runPid := exec.Command("ffmpeg", "-i", fileName, "-r", "1", "-f", "image2", fmt.Sprintf("%s/%s-%%5d.jpg", workerImagePath, msg.id.Hex()))
	start := time.Now()

	errCmd := runPid.Start()

	if errCmd != nil {
		logger.Errorf("execute ffmpeg false with error:%s\n", errCmd.Error())
		return err
	}

	err = runPid.Wait()
	if err != nil {
		logger.Errorf("runPid,Wait() with error:%v\n", err)
		return err
	}
	totalSec := time.Now().Sub(start).Seconds()

	logger.Printf("transfer video to images succeed with total %f seconds\n", totalSec)

	dir, err := ioutil.ReadDir(workerImagePath)

	if err != nil {
		logger.Println("ioutil.ReadDir(%s) with error:%s", workerImagePath, err)
		return err
	}
	formUploader := storage.NewFormUploader(&cfg)

	ret := storage.PutRet{}

	imgs := []string{}

	pubPolicy := storage.PutPolicy{
		Scope:   bucket,
		Expires: 3600 * 3,
	}

	uptoken := pubPolicy.UploadToken(mac)

	imgNum := 0

	for imgIndex, img := range dir {

		if img.IsDir() || !strings.HasPrefix(img.Name(), msg.id.Hex()) || !strings.HasSuffix(img.Name(), "jpg") {
			continue
		}
		imgNum += 1
		logger.Printf("upload %d/%d\n", imgIndex, len(dir))
		err = formUploader.PutFile(context.Background(), &ret, uptoken, img.Name(), path.Join(workerImagePath, img.Name()), nil)

		if err != nil {
			logger.Errorf("formUploader.PutFile(ctx,ret,uptoken,%s,%s,)", img.Name(), path.Join(workerImagePath, img.Name()))
			continue
		}

		imgs = append(imgs, img.Name())

	}

	logger.Println("upload finished")

	sort.Strings(imgs)

	results := []models.ResultBody{}

	for _, choice := range msg.choices {
		if choice != "scene" && choice != "object" && choice != "people" {
			continue
		}
		result := models.ResultBody{Type: choice, Result: []models.Result{}}

		if choice == "people" {
			for i, img := range imgs {
				param := ReqParam{Data: Data{URI: conf.BktHost + "/" + img}}
				jsVal, err := json.Marshal(param)
				if err != nil {
					logger.Errorf("json.Marshal(%+v) with error:%+v\n", param, err)
					continue
				}
				req, err := http.NewRequest("POST", type2APIMap[choice], strings.NewReader(string(jsVal)))
				if err != nil {
					logger.Errorf("http.NewRequest('POST',%s,%v) with error:%v\n", type2APIMap[choice], jsVal, err)
					continue
				}
				mac1 := &qiniumac.Mac{
					AccessKey: conf.AK,
					SecretKey: []byte(conf.SK),
				}

				t := qiniumac.NewTransport(mac1, http.DefaultTransport)
				cli := &http.Client{Transport: t}
				req.Header.Set("Content-Type", "application/json")
				resp, err := cli.Do(req)
				//logger.Info(resp.Status)
				if err != nil || resp == nil {
					logger.Errorf("http.Client.Do() with error:%v\n", err)
					continue
				}
				defer resp.Body.Close()
				if resp != nil && resp.StatusCode != 200 {
					body, _ := ioutil.ReadAll(resp.Body)
					logger.Errorf("HTTP request:%s status:%d\n", type2APIMap[choice], resp.StatusCode)
					logger.Error(string(body))
					continue
				}

				resps := RespBody{}

				body, err := ioutil.ReadAll(resp.Body)
				logger.Info(string(body))

				//logger.Infof("%v", string(body))
				if err != nil {
					logger.Errorf("ioutil.ReadAll() with error:%v\n", body)
					continue
				}

				err = json.Unmarshal(body, &resps)
				resultItem := models.Result{}
				apiResults, ok := resps.Result.(map[string]interface{})["detections"].([]interface{})
				if !ok {
					continue
				}
				for _, apiRes := range apiResults {
					res := apiRes.(map[string]interface{})

					if name, ok := res["name"].(string); ok {
						resultItem.Attribute = name

					} else {
						continue
					}
					resultItem.Confidence = res["score"].(float64)
					resultItem.Type = choice
					resultItem.Time.Start = i
					resultItem.Time.End = i

					result.Result = append(result.Result, resultItem)
				}
			}
			//result.Result = models.MergeResult(result.Result)

		} else {
			for i := 0; i < len(imgs); i += chunkSize {
				end := i + chunkSize
				logger.Info("Enter into", strconv.Itoa(i))
				if end > len(imgs) {
					end = len(imgs)
				}

				params := []BatchReqParam{}

				for _, img := range imgs[i:end] {

					bachParam := BatchReqParam{
						Op:   type2APIMap[choice],
						Data: Data{URI: conf.BktHost + "/" + img},
					}
					params = append(params, bachParam)
				}

				jsonVal, err := json.Marshal(params)

				if err != nil {
					logger.Errorf("json.Marshal(%v) with error:%v\n", params, err)
					continue
				}

				req, err := http.NewRequest("POST", BATCH_API, strings.NewReader(string(jsonVal)))
				if err != nil {
					logger.Errorf("http.NewRequest('POST',%s,%v) with error:%v\n", BATCH_API, jsonVal, err)
					continue
				}
				mac2 := &qiniumac.Mac{
					AccessKey: conf.AK,
					SecretKey: []byte(conf.SK),
				}

				t := qiniumac.NewTransport(mac2, http.DefaultTransport)
				cli := &http.Client{Transport: t}
				req.Header.Set("Content-Type", "application/json")
				resp, err := cli.Do(req)
				//logger.Info(resp.Status)
				if err != nil || resp == nil {
					logger.Errorf("http.Client.Do() with error:%v\n", err)
					continue
				}
				if resp != nil && resp.StatusCode != 200 {
					logger.Errorf("HTTP status:%d\n", resp.StatusCode)
					continue
				}

				defer resp.Body.Close()

				resps := []RespBody{}

				body, err := ioutil.ReadAll(resp.Body)

				//logger.Infof("%v", string(body))
				if err != nil {
					logger.Errorf("ioutil.ReadAll() with error:%v\n", body)
					continue
				}

				err = json.Unmarshal(body, &resps)
				resultItem := models.Result{}

				for j, resp := range resps {
					key := ""
					if choice == "scene" {
						key = "confidences"
					} else if choice == "object" || choice == "people" {
						key = "detections"
					} else {
						continue
					}

					apiResults := resp.Result.(map[string]interface{})[key].([]interface{})
					for _, apiRes := range apiResults {
						res := apiRes.(map[string]interface{})

						resultItem.Attribute = res["class"].(string)
						if val, ok := res["label_cn"]; ok {
							resultItem.Attribute = val.(string)
						}
						resultItem.Confidence = res["score"].(float64)
						resultItem.Type = choice
						resultItem.Time.Start = i + j
						resultItem.Time.End = i + j

						result.Result = append(result.Result, resultItem)
					}
				}
			}
		}

		result.Result = models.MergeResult(result.Result)
		if choice == "scene" {
			result.Result = models.FilterScene(result.Result)
		}
		results = append(results, result)

	}
	logger.Info("task done,update now")

	taskModel, err := models.Task.Find(msg.id.Hex())
	if err != nil {
		return
	}
	taskModel.TotalSecond = imgNum
	taskModel.Results = results
	taskModel.Status = models.TaskStatusDone
	err = taskModel.Save()
	if err != nil {
		return
	}

	if msg.callBack != "" {
		bs, err := json.Marshal(taskModel)
		if err != nil {
			return err
		}
		body := bytes.NewBuffer([]byte(bs))

		resp, err := http.Post(msg.callBack, "application/json;charset=utf-8", body)
		if err != nil || resp == nil {
			return err
		}
		if resp.StatusCode != 200 {
			err = errors.New("call back return status error")
			logger.Errorf("call back id(%s) url:%s failed with response status %d\n", msg.id.Hex(), msg.callBack, resp.StatusCode)
			return err
		}
	}

	logger.Info("task update success!")

	return
}

func download(URI string, dstPath string) (string, error) {
	u, err := url.Parse(URI)
	if err != nil {
		logger.Errorf("url.Parse(%s)", URI)
		return "", errors.New("uri not right")
	}
	var req *http.Request

	switch u.Scheme {
	case "qiniu":

		req, err = http.NewRequest("GET", "http://iovip.qbox.me"+u.Path, nil)
		req.Host = u.Host

	case "http", "https":
		req, err = http.NewRequest("GET", URI, nil)
	}

	if err != nil {
		return "", err
	}
	log.Print("the uri:", URI, dstPath)
	req.Header.Add("User-Agent", "video struct")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		logger.Errorf("http.DefaultClient.do(%v)", req)
		return "", err
	}
	defer resp.Body.Close()

	fileName := path.Join(dstPath, strings.TrimLeft(u.Path, "/"))
	f, err := os.Create(fileName)

	if err != nil {
		logger.Errorf("os.Create(%s) with error:%v", fileName, err)
		return "", err
	}
	defer f.Close()

	io.Copy(f, resp.Body)

	return fileName, err

}

func main() {

	//Load config
	config.Init("f", "", "app.conf")
	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("main.config.Load", err)
	}

	log.SetOutputLevel(conf.DebugLevel)
	runtime.GOMAXPROCS(conf.MaxProcs)

	mac = qbox.NewMac(AK, SK)

	cfg := storage.Config{}

	cfg.Zone = &storage.ZoneHuadong

	cfg.UseHTTPS = false

	cfg.UseCdnDomains = false

	logger = log.New(os.Stdout, "[info]", log.LstdFlags)
	logger.SetOutputLevel(conf.DebugLevel)

	models.SetupModel(model.NewModel(&conf.Mgo, *logger))

	srcPath, err := os.Getwd()
	if err != nil {
		logger.Errorf("error when get current pwd")
		return
	}

	workspace = path.Join(srcPath, "videos")

	if _, err = os.Stat(workspace); os.IsNotExist(err) {
		errMkdir := os.Mkdir(workspace, os.ModePerm)
		if errMkdir != nil {
			logger.Errorf("os.Mkdir(%s,os.ModePerm) with error:%v \n", workspace, errMkdir)
			return
		}
	}

	router := gin.Default()

	cpuNum := runtime.NumCPU()

	for i := 0; i < cpuNum; i++ {
		logger.Printf("start worker[%d]\n", i)

		logger.Printf("create the the worker workspace[%d]\n", i)

		workerIPath := path.Join(workspace, strconv.Itoa(i))
		if _, err = os.Stat(workerIPath); os.IsNotExist(err) {
			errMkdir := os.Mkdir(workerIPath, os.ModePerm)
			if errMkdir != nil {
				logger.Errorf("os.Mkdir(%s,os.ModePerm) with error:%v \n", workerIPath, errMkdir)
				return
			}
		}

		go func(workerPath string, conf *Config) {
			log.Print("the workerpath:", workerPath)

			for msg := range msgsChan {

				if result := do(msg, workerPath, conf); result != nil {
					taskModel, err := models.Task.Find(msg.id.Hex())
					if err != nil {
						logger.Errorf("models.Task.Find(%s) with error:%v\n", msg.id.Hex(), err)
						continue
					}
					taskModel.Status = models.TaskStatusError
					err = taskModel.Save()
					if err != nil {
						logger.Errorf("taskModel(id=%s).Save() failed with error:%v\n", taskModel.Id.Hex(), err)
					}
				}

			}
		}(workerIPath, &conf)
	}

	router.PUT("/v1/video", func(c *gin.Context) {
		var json videoRequest
		if err = c.ShouldBindJSON(&json); err != nil {

			c.JSON(http.StatusNotImplemented, map[string]interface{}{
				"task_id": "null",
				"status":  "create failed",
			})
			return
		}
		task := models.NewTaskModel(json.Src, json.Name, json.Choice)

		job := Job{
			fileURI:  json.Src,
			name:     json.Name,
			choices:  strings.Split(json.Choice, "|"),
			callBack: json.CallBack,
			id:       task.Id,
		}

		msgsChan <- job

		err = task.Save()
		if err != nil {
			c.JSON(http.StatusNotImplemented, map[string]interface{}{
				"task_id": task.Id.Hex(),
				"status":  "create failed",
			})
			return
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"task_id": task.Id.Hex(),
			"status":  "created",
		})

		return

	})

	router.POST("/v1/testvideocallback", func(c *gin.Context) {
		var task models.TaskModel
		if err = c.ShouldBindJSON(&task); err != nil {
			c.JSON(http.StatusOK, map[string]interface{}{})
		}
		logger.Infof("test callback result:%+v\n", task)
		return

	})

	router.GET("/v1/video", func(c *gin.Context) {
		taskId := c.Query("task_id")
		if taskId == "" {
			c.JSON(http.StatusBadRequest, map[string]interface{}{})
			return
		}

		taskModel, err := models.Task.Find(taskId)
		if err != nil {
			logger.Errorf("models.Task.Find(%s) with error:%v\n", taskId, err)
			c.JSON(http.StatusNotFound, nil)
			return
		}

		c.JSON(http.StatusOK, taskModel)
		return
	})

	router.Run(conf.BindHost)

}
