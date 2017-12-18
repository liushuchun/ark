package atfuck

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/mholt/archiver"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"qiniu/api.v6/auth/digest"
	"qiniu/api.v6/conf"
)

/*
{
	"dest_dir"		:	"/Users/jemy/Backup",
	"bucket"		:	"test-bucket",
	"prefix"		:	"demo/",
	"suffixes"		: 	".png,.jpg",
}
*/

const (
	MIN_DOWNLOAD_THREAD_COUNT = 1
	MAX_DOWNLOAD_THREAD_COUNT = 2000
)

type DownloadConfig struct {
	DestDir  string `json:"dest_dir"`
	Bucket   string `json:"bucket"`
	Prefix   string `json:"prefix,omitempty"`
	AK       string `json:"ak"`
	SK       string `json:"sk"`
	Suffixes string `json:"suffixes,omitempty"`
	Workers  int64  `json:"workers,omitempty"`
	UnZip    bool   `json:"unzip,omitempty"`
	UnZipDir string `json:"unzip_dir,omitempty"`
	//down from cdn
	Referer   string `json:"referer,omitempty"`
	CdnDomain string `json:"cdn_domain,omitempty"`
	//log settings
	LogLevel  string `json:"log_level,omitempty"`
	LogFile   string `json:"log_file,omitempty"`
	LogRotate int    `json:"log_rotate,omitempty"`
	LogStdout bool   `json:"log_stdout,omitempty"`
}

var downloadTasks chan func()
var initDownOnce sync.Once
var ZIP_LIST = []string{
	".zip",
	".tar",
	".tar.gz",
	".tgz",
	".tar.bz2",
	".tbz2",
	".tar.xz",
	".txz",
	".tar.lz4",
	".tlz4",
	".tar.sz",
	".tsz",
	".rar",
}

func doDownload(tasks chan func()) {
	for {
		task := <-tasks
		task()
	}
}

func QiniuDownload(threadCount int, downConfig *DownloadConfig) {
	timeStart := time.Now()
	//create job id
	jobId := Md5Hex(fmt.Sprintf("%s:%s", downConfig.DestDir, downConfig.Bucket))

	//local storage path
	storePath := filepath.Join(QShellRootPath, ".atfuck", "qdownload", jobId)
	if mkdirErr := os.MkdirAll(storePath, 0775); mkdirErr != nil {
		logs.Error("Failed to mkdir `%s` due to `%s`", storePath, mkdirErr)
		os.Exit(STATUS_ERROR)
	}

	//init log settings
	defaultLogFile := filepath.Join(storePath, fmt.Sprintf("%s.log", jobId))
	//init log level
	logLevel := logs.LevelInfo
	logRotate := 1
	if downConfig.LogRotate > 0 {
		logRotate = downConfig.LogRotate
	}
	switch downConfig.LogLevel {
	case "debug":
		logLevel = logs.LevelDebug
	case "info":
		logLevel = logs.LevelInfo
	case "warn":
		logLevel = logs.LevelWarning
	case "error":
		logLevel = logs.LevelError
	default:
		logLevel = logs.LevelInfo
	}

	//init log writer
	if downConfig.LogFile == "" {
		//set default log file
		downConfig.LogFile = defaultLogFile
	}

	if !downConfig.LogStdout {
		logs.GetBeeLogger().DelLogger(logs.AdapterConsole)
	}
	//open log file
	fmt.Println("Writing download log to file", downConfig.LogFile)

	//daily rotate
	logCfg := BeeLogConfig{
		Filename: downConfig.LogFile,
		Level:    logLevel,
		Daily:    true,
		MaxDays:  logRotate,
	}
	logs.SetLogger(logs.AdapterFile, logCfg.ToJson())
	fmt.Println()

	mac := digest.Mac{downConfig.AK, []byte(downConfig.SK)}
	//get bucket zone info
	bucketInfo, gErr := GetBucketInfo(&mac, downConfig.Bucket)
	if gErr != nil {
		logs.Error("Get bucket region info error,", gErr)
		os.Exit(STATUS_ERROR)
	}
	//get domains of bucket
	domainsOfBucket, gErr := GetDomainsOfBucket(&mac, downConfig.Bucket)
	if gErr != nil {
		logs.Error("Get domains of bucket error,", gErr)
		os.Exit(STATUS_ERROR)
	}

	if len(domainsOfBucket) == 0 {
		logs.Error("No domains found for bucket", downConfig.Bucket)
		os.Exit(STATUS_ERROR)
	}

	domainOfBucket := domainsOfBucket[0]

	//set up host
	SetZone(bucketInfo.Region)
	ioProxyAddress := conf.IO_HOST

	//check whether cdn domain is set
	if downConfig.CdnDomain != "" {
		ioProxyAddress = downConfig.CdnDomain
	}

	//trim http and https prefix
	ioProxyAddress = strings.TrimPrefix(ioProxyAddress, "http://")
	ioProxyAddress = strings.TrimPrefix(ioProxyAddress, "https://")
	if downConfig.CdnDomain != "" {
		domainOfBucket = ioProxyAddress
	}

	jobListFileName := filepath.Join(storePath, fmt.Sprintf("%s.list", jobId))
	resumeFile := filepath.Join(storePath, fmt.Sprintf("%s.ldb", jobId))
	resumeLevelDb, openErr := leveldb.OpenFile(resumeFile, nil)
	if openErr != nil {
		logs.Error("Open resume record leveldb error", openErr)
		os.Exit(STATUS_ERROR)
	}
	defer resumeLevelDb.Close()
	//sync underlying writes from the OS buffer cache
	//through to actual disk
	ldbWOpt := opt.WriteOptions{
		Sync: true,
	}

	//list bucket, prepare file list to download
	logs.Info("Listing bucket `%s` by prefix `%s`", downConfig.Bucket, downConfig.Prefix)
	listErr := ListBucket(&mac, downConfig.Bucket, downConfig.Prefix, "", jobListFileName)
	if listErr != nil {
		logs.Error("List bucket error", listErr)
		os.Exit(STATUS_ERROR)
	}

	//init wait group
	downWaitGroup := sync.WaitGroup{}

	initDownOnce.Do(func() {
		downloadTasks = make(chan func(), threadCount)
		for i := 0; i < threadCount; i++ {
			go doDownload(downloadTasks)
		}
	})

	//init counters
	var totalFileCount int64
	var currentFileCount int64
	var existsFileCount int64
	var updateFileCount int64
	var successFileCount int64
	var failureFileCount int64
	var skipBySuffixes int64

	totalFileCount = GetFileLineCount(jobListFileName)

	//open prepared file list to download files
	listFp, openErr := os.Open(jobListFileName)
	if openErr != nil {
		logs.Error("Open list file error", openErr)
		os.Exit(STATUS_ERROR)
	}
	defer listFp.Close()

	listScanner := bufio.NewScanner(listFp)
	listScanner.Split(bufio.ScanLines)
	//key, fsize, etag, lmd, mime, enduser

	downSuffixes := strings.Split(downConfig.Suffixes, ",")
	filterSuffixes := make([]string, 0, len(downSuffixes))

	for _, suffix := range downSuffixes {
		if strings.TrimSpace(suffix) != "" {
			filterSuffixes = append(filterSuffixes, suffix)
		}
	}

	for listScanner.Scan() {
		currentFileCount += 1
		line := strings.TrimSpace(listScanner.Text())
		items := strings.Split(line, "\t")
		if len(items) >= 4 {
			fileKey := items[0]

			if len(filterSuffixes) > 0 {
				//filter files by suffixes
				var goAhead bool
				for _, suffix := range filterSuffixes {
					if strings.HasSuffix(fileKey, suffix) {
						goAhead = true
						break
					}
				}

				if !goAhead {
					skipBySuffixes += 1
					logs.Info("Skip download `%s`, suffix filter not match", fileKey)
					continue
				}
			}

			fileSize, pErr := strconv.ParseInt(items[1], 10, 64)
			if pErr != nil {
				logs.Error("Invalid list line", line)
				continue
			}

			fileMtime, pErr := strconv.ParseInt(items[3], 10, 64)
			if pErr != nil {
				logs.Error("Invalid list line", line)
				continue
			}

			fileUrl := makePrivateDownloadLink(&mac, domainOfBucket, ioProxyAddress, fileKey)

			//progress
			if totalFileCount != 0 {
				fmt.Printf("Downloading %s [%d/%d, %.1f%%] ...\n", fileKey, currentFileCount, totalFileCount,
					float32(currentFileCount)*100/float32(totalFileCount))
			} else {
				fmt.Printf("Downloading %s ...\n", fileKey)
			}
			//check whether log file exists
			localFilePath := filepath.Join(downConfig.DestDir, fileKey)
			localAbsFilePath, _ := filepath.Abs(localFilePath)
			localFilePathTmp := fmt.Sprintf("%s.tmp", localFilePath)
			localFileInfo, statErr := os.Stat(localFilePath)

			var downNewFile bool
			var fromBytes int64

			if statErr == nil {
				//log file exists, check whether have updates
				oldFileInfo, notFoundErr := resumeLevelDb.Get([]byte(localFilePath), nil)
				if notFoundErr == nil {
					//if exists
					oldFileInfoItems := strings.Split(string(oldFileInfo), "|")
					oldFileLmd, _ := strconv.ParseInt(oldFileInfoItems[0], 10, 64)
					//oldFileSize, _ := strconv.ParseInt(oldFileInfoItems[1], 10, 64)
					if oldFileLmd == fileMtime && localFileInfo.Size() == fileSize {
						//nothing change, ignore
						logs.Info("Local file `%s` exists, same as in bucket, download skip", localAbsFilePath)
						existsFileCount += 1
						continue
					} else {
						//somthing changed, must download a new file
						logs.Info("Local file `%s` exists, but remote file changed, go to download", localAbsFilePath)
						downNewFile = true
					}
				} else {
					if localFileInfo.Size() != fileSize {
						logs.Info("Local file `%s` exists, size not the same as in bucket, go to download", localAbsFilePath)
						downNewFile = true
					} else {
						//treat the local file not changed, write to leveldb, though may not accurate
						//nothing to do
						logs.Warning("Local file `%s` exists with same size as `%s`, treat it not changed", localAbsFilePath, fileKey)
						atomic.AddInt64(&existsFileCount, 1)
						continue
					}
				}
			} else {
				//check whether tmp file exists
				localTmpFileInfo, statErr := os.Stat(localFilePathTmp)
				if statErr == nil {
					//if tmp file exists, check whether last modify changed
					oldFileInfo, notFoundErr := resumeLevelDb.Get([]byte(localFilePath), nil)
					if notFoundErr == nil {
						//if exists
						oldFileInfoItems := strings.Split(string(oldFileInfo), "|")
						oldFileLmd, _ := strconv.ParseInt(oldFileInfoItems[0], 10, 64)
						//oldFileSize, _ := strconv.ParseInt(oldFileInfoItems[1], 10, 64)
						if oldFileLmd == fileMtime {
							//tmp file exists, file not changed, use range to download
							if localTmpFileInfo.Size() < fileSize {
								fromBytes = localTmpFileInfo.Size()
							} else {
								//rename it
								renameErr := os.Rename(localFilePathTmp, localFilePath)
								if renameErr != nil {
									logs.Error("Rename temp file `%s` to final file `%s` error", localFilePathTmp, localFilePath, renameErr)
								}
								continue
							}
						} else {
							logs.Info("Local tmp file `%s` exists, but remote file changed, go to download", localFilePathTmp)
							downNewFile = true
						}
					} else {
						//log tmp file exists, but no record in leveldb, download a new log file
						logs.Info("Local tmp file `%s` exists, but no record in leveldb ,go to download", localFilePathTmp)
						downNewFile = true
					}
				} else {
					//no log file exists, donwload a new log file
					downNewFile = true
				}
			}

			//set file info in leveldb
			rKey := localAbsFilePath
			rVal := fmt.Sprintf("%d|%d", fileMtime, fileSize)
			resumeLevelDb.Put([]byte(rKey), []byte(rVal), &ldbWOpt)

			//download new
			downWaitGroup.Add(1)
			downloadTasks <- func() {
				defer downWaitGroup.Done()

				downErr := downloadFile(downConfig, fileKey, fileUrl, domainOfBucket, fileSize, fromBytes)
				if downErr != nil {
					atomic.AddInt64(&failureFileCount, 1)
					logs.Info("put into the queue again")

				} else {
					atomic.AddInt64(&successFileCount, 1)
					if !downNewFile {
						atomic.AddInt64(&updateFileCount, 1)
					}
				}
			}
		}
	}

	//wait for all tasks done
	downWaitGroup.Wait()

	logs.Info("-------Download Result-------")
	logs.Info("%10s%10d", "Total:", totalFileCount)
	logs.Info("%10s%10d", "Skipped:", skipBySuffixes)
	logs.Info("%10s%10d", "Exists:", existsFileCount)
	logs.Info("%10s%10d", "Success:", successFileCount)
	logs.Info("%10s%10d", "Update:", updateFileCount)
	logs.Info("%10s%10d", "Failure:", failureFileCount)
	logs.Info("%10s%15s", "Duration:", time.Since(timeStart))
	logs.Info("-----------------------------")
	fmt.Println("\nSee download log at path", downConfig.LogFile)

	if failureFileCount > 0 {
		os.Exit(STATUS_ERROR)
	}
}

/*
@param ioProxyHost - like http://iovip.qbox.me
*/
func makePrivateDownloadLink(mac *digest.Mac, domainOfBucket, ioProxyAddress, fileKey string) (fileUrl string) {
	publicUrl := fmt.Sprintf("http://%s/%s", domainOfBucket, fileKey)
	deadline := time.Now().Add(time.Hour * 24 * 30).Unix()
	privateUrl := PrivateUrl(mac, publicUrl, deadline)

	//replace the io proxy host
	fileUrl = strings.Replace(privateUrl, domainOfBucket, ioProxyAddress, -1)
	return
}

//file key -> mtime
func downloadFile(downConfig *DownloadConfig, fileName, fileUrl, domainOfBucket string, fileSize int64, fromBytes int64) (err error) {
	startDown := time.Now().Unix()
	destDir := downConfig.DestDir
	localFilePath := filepath.Join(destDir, fileName)
	localFileDir := filepath.Dir(localFilePath)
	localFilePathTmp := fmt.Sprintf("%s.tmp", localFilePath)

	mkdirErr := os.MkdirAll(localFileDir, 0775)
	if mkdirErr != nil {
		err = mkdirErr
		logs.Error("MkdirAll failed for", localFileDir, mkdirErr)
		return
	}

	logs.Info("Downloading", fileName, "=>", localFilePath)
	//new request
	req, reqErr := http.NewRequest("GET", fileUrl, nil)
	if reqErr != nil {
		err = reqErr
		logs.Info("New request", fileName, "failed by url", fileUrl, reqErr)
		return
	}
	//set host
	req.Host = domainOfBucket
	if downConfig.Referer != "" {
		req.Header.Add("Referer", downConfig.Referer)
	}

	if fromBytes != 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", fromBytes))
	}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = respErr
		logs.Info("Download", fileName, "failed by url", fileUrl, respErr)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		var localFp *os.File
		var openErr error
		if fromBytes != 0 {
			localFp, openErr = os.OpenFile(localFilePathTmp, os.O_APPEND|os.O_WRONLY, 0655)
		} else {
			localFp, openErr = os.Create(localFilePathTmp)
		}

		if openErr != nil {
			err = openErr
			logs.Error("Open local file", localFilePathTmp, "failed", openErr)
			return
		}

		cpCnt, cpErr := io.Copy(localFp, resp.Body)
		if cpErr != nil {
			err = cpErr
			localFp.Close()
			logs.Error("Download", fileName, "failed", cpErr)
			return
		}
		localFp.Close()
		if cpCnt != fileSize {
			logs.Error("size not equal,cpCnt:%d,fileSize:%d", cpCnt, fileSize)
		}
		endDown := time.Now().Unix()
		avgSpeed := fmt.Sprintf("%.2fKB/s", float64(cpCnt)/float64(endDown-startDown)/1024)

		//move temp file to log file
		renameErr := os.Rename(localFilePathTmp, localFilePath)
		if renameErr != nil {
			err = renameErr
			logs.Error("Rename temp file to final log file error", renameErr)
			return
		}
		logs.Info("Download", fileName, "=>", localFilePath, "success", avgSpeed)

		if downConfig.UnZip {
			/*
				destTarDir := downConfig.UnZipDir
				if destTarDir == "" {
					destTarDir = filepath.Join(destDir, "tar")
				}
				if _, err := os.Stat(destTarDir); err != nil && os.IsNotExist(err) {
					if errMk := os.Mkdir(destDir, os.ModePerm); errMk != nil {
						logs.Error("os.Mkdir(%s):%v", destDir, err)
					}
				}
			*/
			if IsZiped(fileName) {
				if archiver.Zip.Match(localFilePath) {
					if errUnzip := archiver.Zip.Open(localFilePath, localFileDir); errUnzip != nil {
						logs.Error("archiver.Zip.Open(%s,%s):%v", localFilePath, localFileDir, errUnzip)
					} else {
						logs.Info("unzip %s => %s succeed!", localFilePath, localFileDir)
					}
				} else if archiver.Rar.Match(localFilePath) {
					if errUnzip := archiver.Rar.Open(localFilePath, localFileDir); errUnzip != nil {
						logs.Error("archiver.Rar.Open(%s,%s):%v", localFilePath, localFileDir, errUnzip)
					} else {
						logs.Info("rar %s => %s succeed!", localFilePath, localFileDir)
					}
				} else if archiver.TarBz2.Match(localFilePath) {
					if errUnzip := archiver.TarBz2.Open(localFilePath, localFileDir); errUnzip != nil {
						logs.Error("archiver.TarBz2.Open(%s,%s):%v", localFilePath, localFileDir, errUnzip)
					} else {
						logs.Info("tarbz2 %s => %s succeed!", localFilePath, localFileDir)
					}
				} else if archiver.Tar.Match(localFilePath) {
					if errUnzip := archiver.Tar.Open(localFilePath, localFileDir); errUnzip != nil {
						logs.Error("archiver.Tar.Open(%s,%s):%v", localFilePath, localFileDir, errUnzip)
					} else {
						logs.Info("tar %s => %s succeed!", localFilePath, localFileDir)
					}
				} else if archiver.TarGz.Match(localFilePath) {
					if errUnzip := archiver.TarGz.Open(localFilePath, localFileDir); errUnzip != nil {
						logs.Error("archiver.Tar.Open(%s,%s):%v", localFilePath, localFileDir, errUnzip)
					} else {
						logs.Info("targz %s => %s succeed!", localFilePath, localFileDir)
					}
				} else if archiver.TarXZ.Match(localFilePath) {
					if errUnzip := archiver.TarXZ.Open(localFilePath, localFileDir); errUnzip != nil {
						logs.Error("archiver.Tarxz.Open(%s,%s):%v", localFilePath, localFileDir, errUnzip)
					} else {
						logs.Info("tarxz %s => %s succeed!", localFilePath, localFileDir)
					}
				}

			}
		}
	} else {
		err = errors.New("download failed")
		logs.Info("Download", fileName, "failed by url", fileUrl, resp.Status)
		return
	}
	return
}

func IsZiped(fileName string) bool {
	for _, zFile := range ZIP_LIST {
		if strings.HasSuffix(fileName, zFile) {
			return true
		}
	}
	return false
}
