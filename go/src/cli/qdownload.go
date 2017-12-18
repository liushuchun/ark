package cli

import (
	"atfuck"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/astaxie/beego/logs"
)

func QiniuDownload(cmd string, params ...string) {
	if len(params) == 1 || len(params) == 2 {
		var threadCount int64 = 5
		var downloadConfigFile string
		var err error
		if len(params) == 1 {
			downloadConfigFile = params[0]
		} else {
			threadCount, err = strconv.ParseInt(params[0], 10, 64)
			if err != nil {
				logs.Error("Invalid value for <ThreadCount>", params[0])
				os.Exit(atfuck.STATUS_HALT)
			}
			downloadConfigFile = params[1]
		}

		//read download config
		fp, err := os.Open(downloadConfigFile)
		if err != nil {
			logs.Error("Open download config file `%s` error, %s", downloadConfigFile, err)
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		configData, err := ioutil.ReadAll(fp)
		if err != nil {
			logs.Error("Read download config file `%s` error, %s", downloadConfigFile, err)
			os.Exit(atfuck.STATUS_HALT)
		}

		var downloadConfig atfuck.DownloadConfig
		err = json.Unmarshal(configData, &downloadConfig)
		if err != nil {
			logs.Error("Parse download config file `%s` error, %s", downloadConfigFile, err)
			os.Exit(atfuck.STATUS_HALT)
		}

		destFileInfo, err := os.Stat(downloadConfig.DestDir)

		if err != nil {
			logs.Error("Download config error for parameter `DestDir`,", err)
			os.Exit(atfuck.STATUS_HALT)
		}

		if !destFileInfo.IsDir() {
			logs.Error("Download dest dir should be a directory")
			os.Exit(atfuck.STATUS_HALT)
		}
		threadCount = downloadConfig.Workers

		if threadCount < atfuck.MIN_DOWNLOAD_THREAD_COUNT || threadCount > atfuck.MAX_DOWNLOAD_THREAD_COUNT {
			logs.Info("Tip: you can set <ThreadCount> value between %d and %d to improve speed\n",
				atfuck.MIN_DOWNLOAD_THREAD_COUNT, atfuck.MAX_DOWNLOAD_THREAD_COUNT)

			if threadCount < atfuck.MIN_DOWNLOAD_THREAD_COUNT {
				threadCount = atfuck.MIN_DOWNLOAD_THREAD_COUNT
			} else if threadCount > atfuck.MAX_DOWNLOAD_THREAD_COUNT {
				threadCount = atfuck.MAX_DOWNLOAD_THREAD_COUNT
			}
		}

		atfuck.QiniuDownload(int(threadCount), &downloadConfig)
	} else {
		CmdHelp(cmd)
	}
}
