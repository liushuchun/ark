package cli

import (
	"atfuck"
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego/logs"
	"qiniu/api.v6/auth/digest"
)

func Sync(cmd string, params ...string) {
	if len(params) == 3 || len(params) == 4 {
		srcResUrl := params[0]
		bucket := params[1]
		key := params[2]
		upHostIp := ""
		if len(params) == 4 {
			upHostIp = params[3]
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			logs.Error(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		//get bucket zone info
		bucketInfo, gErr := atfuck.GetBucketInfo(&mac, bucket)
		if gErr != nil {
			fmt.Println("Get bucket region info error,", gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		//set up host
		atfuck.SetZone(bucketInfo.Region)

		//sync
		tStart := time.Now()
		syncRet, sErr := atfuck.Sync(&mac, srcResUrl, bucket, key, upHostIp)
		if sErr != nil {
			logs.Error(sErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		fmt.Printf("Sync %s => %s:%s Success, Duration: %s!\n", srcResUrl, bucket, key, time.Since(tStart))
		fmt.Println("Hash:", syncRet.Hash)
		fmt.Printf("Fsize: %d (%s)\n", syncRet.Fsize, FormatFsize(syncRet.Fsize))
		fmt.Println("Mime:", syncRet.MimeType)
	} else {
		CmdHelp(cmd)
	}
}
