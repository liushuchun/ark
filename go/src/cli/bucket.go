package cli

import (
	"atfuck"
	"fmt"
	"os"

	"github.com/astaxie/beego/logs"
	"qiniu/api.v6/auth/digest"
)

func GetBuckets(cmd string, params ...string) {
	if len(params) == 0 {
		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}
		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		buckets, err := atfuck.GetBuckets(&mac)
		if err != nil {
			logs.Error("Get buckets error,", err)
			os.Exit(atfuck.STATUS_ERROR)
		} else {
			if len(buckets) == 0 {
				fmt.Println("No buckets found")
			} else {
				for _, bucket := range buckets {
					fmt.Println(bucket)
				}
			}
		}
	} else {
		CmdHelp(cmd)
	}
}

func GetDomainsOfBucket(cmd string, params ...string) {
	if len(params) == 1 {
		bucket := params[0]
		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}
		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		domains, err := atfuck.GetDomainsOfBucket(&mac, bucket)
		if err != nil {
			logs.Error("Get domains error,", err)
			os.Exit(atfuck.STATUS_ERROR)
		} else {
			if len(domains) == 0 {
				fmt.Printf("No domains found for bucket `%s`\n", bucket)
			} else {
				for _, domain := range domains {
					fmt.Println(domain)
				}
			}
		}
	} else {
		CmdHelp(cmd)
	}
}

func GetFileFromBucket(cmd string, params ...string) {

}
