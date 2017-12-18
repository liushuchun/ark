package cli

import (
	"atfuck"
	"bufio"
	"flag"
	"fmt"
	"os"
	"qiniu/rpc"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"qiniu/api.v6/auth/digest"
	"qiniu/api.v6/rs"
)

const (
	BATCH_ALLOW_MAX = 1000
)

func doBatchOperation(tasks chan func()) {
	for {
		task := <-tasks
		task()
	}
}

func printStat(bucket string, key string, entry rs.Entry) {
	statInfo := fmt.Sprintf("%-20s%s\r\n", "Bucket:", bucket)
	statInfo += fmt.Sprintf("%-20s%s\r\n", "Key:", key)
	statInfo += fmt.Sprintf("%-20s%s\r\n", "Hash:", entry.Hash)
	statInfo += fmt.Sprintf("%-20s%d -> %s\r\n", "Fsize:", entry.Fsize, FormatFsize(entry.Fsize))

	putTime := time.Unix(0, entry.PutTime*100)
	statInfo += fmt.Sprintf("%-20s%d -> %s\r\n", "PutTime:", entry.PutTime, putTime.String())
	statInfo += fmt.Sprintf("%-20s%s\r\n", "MimeType:", entry.MimeType)
	if entry.FileType == 0 {
		statInfo += fmt.Sprintf("%-20s%d -> 标准存储\r\n", "FileType:", entry.FileType)
	} else {
		statInfo += fmt.Sprintf("%-20s%d -> 低频存储\r\n", "FileType:", entry.FileType)
	}
	fmt.Println(statInfo)
}

func DirCache(cmd string, params ...string) {
	if len(params) == 2 {
		cacheRootPath := params[0]
		cacheResultFile := params[1]
		_, retErr := atfuck.DirCache(cacheRootPath, cacheResultFile)
		if retErr != nil {
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func ListBucket(cmd string, params ...string) {
	var listMarker string
	flagSet := flag.NewFlagSet("listbucket", flag.ExitOnError)
	flagSet.StringVar(&listMarker, "marker", "", "list marker")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 2 || len(cmdParams) == 3 {
		bucket := cmdParams[0]
		prefix := ""
		listResultFile := ""
		if len(cmdParams) == 2 {
			listResultFile = cmdParams[1]
		} else if len(cmdParams) == 3 {
			prefix = cmdParams[1]
			listResultFile = cmdParams[2]
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{account.AccessKey, []byte(account.SecretKey)}

		retErr := atfuck.ListBucket(&mac, bucket, prefix, listMarker, listResultFile)
		if retErr != nil {
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Stat(cmd string, params ...string) {
	if len(params) == 2 {
		bucket := params[0]
		key := params[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		client := rs.NewMac(&mac)
		entry, err := client.Stat(nil, bucket, key)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Stat error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Stat error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		} else {
			printStat(bucket, key, entry)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Delete(cmd string, params ...string) {
	if len(params) == 2 {
		bucket := params[0]
		key := params[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		client := rs.NewMac(&mac)
		err := client.Delete(nil, bucket, key)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Delete error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Delete error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Move(cmd string, params ...string) {
	var overwrite bool
	flagSet := flag.NewFlagSet("move", flag.ExitOnError)
	flagSet.BoolVar(&overwrite, "overwrite", false, "overwrite mode")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 3 || len(cmdParams) == 4 {
		srcBucket := cmdParams[0]
		srcKey := cmdParams[1]
		destBucket := cmdParams[2]
		destKey := srcKey
		if len(cmdParams) == 4 {
			destKey = cmdParams[3]
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		client := rs.NewMac(&mac)
		err := client.Move(nil, srcBucket, srcKey, destBucket, destKey, overwrite)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Move error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Move error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Copy(cmd string, params ...string) {
	var overwrite bool
	flagSet := flag.NewFlagSet("copy", flag.ExitOnError)
	flagSet.BoolVar(&overwrite, "overwrite", false, "overwrite mode")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 3 || len(cmdParams) == 4 {
		srcBucket := cmdParams[0]
		srcKey := cmdParams[1]
		destBucket := cmdParams[2]
		destKey := srcKey
		if len(cmdParams) == 4 {
			destKey = cmdParams[3]
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		client := rs.NewMac(&mac)
		err := client.Copy(nil, srcBucket, srcKey, destBucket, destKey, overwrite)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Copy error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Copy error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Chgm(cmd string, params ...string) {
	if len(params) == 3 {
		bucket := params[0]
		key := params[1]
		newMimeType := params[2]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		client := rs.NewMac(&mac)
		err := client.ChangeMime(nil, bucket, key, newMimeType)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Change mimetype error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Change mimetype error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Fetch(cmd string, params ...string) {
	if len(params) == 2 || len(params) == 3 {
		remoteResUrl := params[0]
		bucket := params[1]
		key := ""
		if len(params) == 3 {
			key = params[2]
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		fetchResult, err := atfuck.Fetch(&mac, remoteResUrl, bucket, key)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Fetch error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Fetch error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		} else {
			fmt.Println("Key:", fetchResult.Key)
			fmt.Println("Hash:", fetchResult.Hash)
			fmt.Printf("Fsize: %d (%s)\n", fetchResult.Fsize, FormatFsize(fetchResult.Fsize))
			fmt.Println("Mime:", fetchResult.MimeType)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Prefetch(cmd string, params ...string) {
	if len(params) == 2 {
		bucket := params[0]
		key := params[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		err := atfuck.Prefetch(&mac, bucket, key)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Prefetch error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Prefetch error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}

func BatchStat(cmd string, params ...string) {
	if len(params) == 2 {
		bucket := params[0]
		keyListFile := params[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		client := rs.NewMac(&mac)
		fp, err := os.Open(keyListFile)
		if err != nil {
			fmt.Println("Open key list file error", err)
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		entries := make([]rs.EntryPath, 0, BATCH_ALLOW_MAX)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, "\t")
			if len(items) > 0 {
				key := items[0]
				if key != "" {
					entry := rs.EntryPath{
						bucket, key,
					}
					entries = append(entries, entry)
				}
			}
			//check 1000 limit
			if len(entries) == BATCH_ALLOW_MAX {
				batchStat(client, entries)
				//reset slice
				entries = make([]rs.EntryPath, 0)
			}
		}
		//stat the last batch
		if len(entries) > 0 {
			batchStat(client, entries)
		}
	} else {
		CmdHelp(cmd)
	}
}

func batchStat(client rs.Client, entries []rs.EntryPath) {
	ret, err := atfuck.BatchStat(client, entries)
	if len(ret) > 0 {
		for i, entry := range entries {
			item := ret[i]
			if item.Code != 200 || item.Data.Error != "" {
				fmt.Println(entry.Key + "\t" + item.Data.Error)
			} else {
				fmt.Println(fmt.Sprintf("%s\t%d\t%s\t%s\t%d\t%d", entry.Key,
					item.Data.Fsize, item.Data.Hash, item.Data.MimeType, item.Data.PutTime, item.Data.FileType))
			}
		}
	} else {
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Batch stat error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Batch stat error,", err)
			}
		}
	}
}

func BatchDelete(cmd string, params ...string) {
	var force bool
	var worker int
	flagSet := flag.NewFlagSet("batchdelete", flag.ExitOnError)
	flagSet.BoolVar(&force, "force", false, "force mode")
	flagSet.IntVar(&worker, "worker", 1, "worker count")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 2 {
		if !force {
			//confirm
			rcode := CreateRandString(6)

			rcode2 := ""
			if runtime.GOOS == "windows" {
				fmt.Print(fmt.Sprintf("<DANGER> Input %s to confirm operation: ", rcode))
			} else {
				fmt.Print(fmt.Sprintf("\033[31m<DANGER>\033[0m Input \033[32m%s\033[0m to confirm operation: ", rcode))
			}
			fmt.Scanln(&rcode2)

			if rcode != rcode2 {
				fmt.Println("Task quit!")
				os.Exit(atfuck.STATUS_HALT)
			}
		}

		bucket := cmdParams[0]
		keyListFile := cmdParams[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		//set zone info
		atfuck.SetZone(bucketInfo.Region)

		var batchTasks chan func()
		var initBatchOnce sync.Once

		batchWaitGroup := sync.WaitGroup{}
		initBatchOnce.Do(func() {
			batchTasks = make(chan func(), worker)
			for i := 0; i < worker; i++ {
				go doBatchOperation(batchTasks)
			}
		})

		client := rs.NewMac(&mac)
		fp, err := os.Open(keyListFile)
		if err != nil {
			fmt.Println("Open key list file error", err)
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		entries := make([]rs.EntryPath, 0, BATCH_ALLOW_MAX)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, "\t")
			if len(items) > 0 {
				key := items[0]
				if key != "" {
					entry := rs.EntryPath{
						bucket, key,
					}
					entries = append(entries, entry)
				}
			}
			//check limit
			if len(entries) == BATCH_ALLOW_MAX {
				toDeleteEntries := make([]rs.EntryPath, len(entries))
				copy(toDeleteEntries, entries)

				batchWaitGroup.Add(1)
				batchTasks <- func() {
					defer batchWaitGroup.Done()
					batchDelete(client, toDeleteEntries)
				}
				entries = make([]rs.EntryPath, 0, BATCH_ALLOW_MAX)
			}
		}
		//delete the last batch
		if len(entries) > 0 {
			toDeleteEntries := make([]rs.EntryPath, len(entries))
			copy(toDeleteEntries, entries)

			batchWaitGroup.Add(1)
			batchTasks <- func() {
				defer batchWaitGroup.Done()
				batchDelete(client, toDeleteEntries)
			}
		}

		batchWaitGroup.Wait()
	} else {
		CmdHelp(cmd)
	}
}

func batchDelete(client rs.Client, entries []rs.EntryPath) {
	ret, err := atfuck.BatchDelete(client, entries)

	if len(ret) > 0 {
		for i, entry := range entries {
			item := ret[i]

			if item.Code != 200 || item.Data.Error != "" {
				logs.Error("Delete '%s' => '%s' failed, Code: %d, Error: %s", entry.Bucket, entry.Key, item.Code, item.Data.Error)
			} else {
				logs.Debug("Delete '%s' => '%s' success", entry.Bucket, entry.Key)
			}
		}
	} else {
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Batch delete error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Batch delete error,", err)
			}
		}
	}
}

func BatchChgm(cmd string, params ...string) {
	var force bool
	var worker int
	flagSet := flag.NewFlagSet("batchchgm", flag.ExitOnError)
	flagSet.BoolVar(&force, "force", false, "force mode")
	flagSet.IntVar(&worker, "worker", 1, "worker count")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 2 {
		if !force {
			//confirm
			rcode := CreateRandString(6)

			rcode2 := ""
			if runtime.GOOS == "windows" {
				fmt.Printf("<DANGER> Input %s to confirm operation: ", rcode)
			} else {
				fmt.Printf("\033[31m<DANGER>\033[0m Input \033[32m%s\033[0m to confirm operation: ", rcode)
			}
			fmt.Scanln(&rcode2)

			if rcode != rcode2 {
				fmt.Println("Task quit!")
				os.Exit(atfuck.STATUS_HALT)
			}
		}

		bucket := cmdParams[0]
		keyMimeMapFile := cmdParams[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		//set zone info
		atfuck.SetZone(bucketInfo.Region)

		var batchTasks chan func()
		var initBatchOnce sync.Once

		batchWaitGroup := sync.WaitGroup{}
		initBatchOnce.Do(func() {
			batchTasks = make(chan func(), worker)
			for i := 0; i < worker; i++ {
				go doBatchOperation(batchTasks)
			}
		})

		client := rs.NewMac(&mac)
		fp, err := os.Open(keyMimeMapFile)
		if err != nil {
			fmt.Println("Open key mime map file error")
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		entries := make([]atfuck.ChgmEntryPath, 0, BATCH_ALLOW_MAX)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, "\t")
			if len(items) == 2 {
				key := items[0]
				mimeType := items[1]
				if key != "" && mimeType != "" {
					entry := atfuck.ChgmEntryPath{bucket, key, mimeType}
					entries = append(entries, entry)
				}
			}
			if len(entries) == BATCH_ALLOW_MAX {
				toChgmEntries := make([]atfuck.ChgmEntryPath, len(entries))
				copy(toChgmEntries, entries)

				batchWaitGroup.Add(1)
				batchTasks <- func() {
					defer batchWaitGroup.Done()
					batchChgm(client, toChgmEntries)
				}
				entries = make([]atfuck.ChgmEntryPath, 0, BATCH_ALLOW_MAX)
			}
		}
		if len(entries) > 0 {
			toChgmEntries := make([]atfuck.ChgmEntryPath, len(entries))
			copy(toChgmEntries, entries)

			batchWaitGroup.Add(1)
			batchTasks <- func() {
				defer batchWaitGroup.Done()
				batchChgm(client, toChgmEntries)
			}
		}

		batchWaitGroup.Wait()
	} else {
		CmdHelp(cmd)
	}
}

func batchChgm(client rs.Client, entries []atfuck.ChgmEntryPath) {
	ret, err := atfuck.BatchChgm(client, entries)
	if len(ret) > 0 {
		for i, entry := range entries {
			item := ret[i]
			if item.Code != 200 || item.Data.Error != "" {
				logs.Error("Chgm '%s' => '%s' Failed, Code: %d, Error: %s", entry.Key, entry.MimeType, item.Code, item.Data.Error)
			} else {
				logs.Debug("Chgm '%s' => '%s' success", entry.Key, entry.MimeType)
			}
		}
	} else {
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Batch chgm error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Batch chgm error,", err)
			}
		}
	}
}

func BatchRename(cmd string, params ...string) {
	var force bool
	var overwrite bool
	var worker int
	flagSet := flag.NewFlagSet("batchrename", flag.ExitOnError)
	flagSet.BoolVar(&force, "force", false, "force mode")
	flagSet.BoolVar(&overwrite, "overwrite", false, "overwrite mode")
	flagSet.IntVar(&worker, "worker", 1, "worker count")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 2 {
		if !force {
			//confirm
			rcode := CreateRandString(6)

			rcode2 := ""
			if runtime.GOOS == "windows" {
				fmt.Printf("<DANGER> Input %s to confirm operation: ", rcode)
			} else {
				fmt.Printf("\033[31m<DANGER>\033[0m Input \033[32m%s\033[0m to confirm operation: ", rcode)
			}
			fmt.Scanln(&rcode2)

			if rcode != rcode2 {
				fmt.Println("Task quit!")
				os.Exit(atfuck.STATUS_HALT)
			}
		}

		bucket := cmdParams[0]
		oldNewKeyMapFile := cmdParams[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		//set zone info
		atfuck.SetZone(bucketInfo.Region)

		var batchTasks chan func()
		var initBatchOnce sync.Once

		batchWaitGroup := sync.WaitGroup{}
		initBatchOnce.Do(func() {
			batchTasks = make(chan func(), worker)
			for i := 0; i < worker; i++ {
				go doBatchOperation(batchTasks)
			}
		})

		client := rs.NewMac(&mac)
		fp, err := os.Open(oldNewKeyMapFile)
		if err != nil {
			fmt.Println("Open old new key map file error")
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		entries := make([]atfuck.RenameEntryPath, 0, BATCH_ALLOW_MAX)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, "\t")
			if len(items) == 2 {
				oldKey := items[0]
				newKey := items[1]
				if oldKey != "" && newKey != "" {
					entry := atfuck.RenameEntryPath{bucket, oldKey, newKey}
					entries = append(entries, entry)
				}
			}
			if len(entries) == BATCH_ALLOW_MAX {
				toRenameEntries := make([]atfuck.RenameEntryPath, len(entries))
				copy(toRenameEntries, entries)

				batchWaitGroup.Add(1)
				batchTasks <- func() {
					defer batchWaitGroup.Done()
					batchRename(client, toRenameEntries, overwrite)
				}
				entries = make([]atfuck.RenameEntryPath, 0, BATCH_ALLOW_MAX)
			}
		}
		if len(entries) > 0 {
			toRenameEntries := make([]atfuck.RenameEntryPath, len(entries))
			copy(toRenameEntries, entries)

			batchWaitGroup.Add(1)
			batchTasks <- func() {
				defer batchWaitGroup.Done()
				batchRename(client, toRenameEntries, overwrite)
			}
		}
		batchWaitGroup.Wait()
	} else {
		CmdHelp(cmd)
	}
}

func batchRename(client rs.Client, entries []atfuck.RenameEntryPath, overwrite bool) {
	ret, err := atfuck.BatchRename(client, entries, overwrite)

	if len(ret) > 0 {
		for i, entry := range entries {
			item := ret[i]
			if item.Code != 200 || item.Data.Error != "" {
				logs.Error("Rename '%s' => '%s' Failed, Code: %d, Error: %s", entry.OldKey, entry.NewKey, item.Code, item.Data.Error)
			} else {
				logs.Debug("Rename '%s' => '%s' success", entry.OldKey, entry.NewKey)
			}
		}
	} else {
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Batch rename error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Batch rename error,", err)
			}
		}
	}
}

func BatchMove(cmd string, params ...string) {
	var force bool
	var overwrite bool
	var worker int
	flagSet := flag.NewFlagSet("batchmove", flag.ExitOnError)
	flagSet.BoolVar(&force, "force", false, "force mode")
	flagSet.BoolVar(&overwrite, "overwrite", false, "overwrite mode")
	flagSet.IntVar(&worker, "worker", 1, "worker count")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 3 {
		if !force {
			//confirm
			rcode := CreateRandString(6)

			rcode2 := ""
			if runtime.GOOS == "windows" {
				fmt.Printf("<DANGER> Input %s to confirm operation: ", rcode)
			} else {
				fmt.Printf("\033[31m<DANGER>\033[0m Input \033[32m%s\033[0m to confirm operation: ", rcode)
			}
			fmt.Scanln(&rcode2)

			if rcode != rcode2 {
				fmt.Println("Task quit!")
				os.Exit(atfuck.STATUS_HALT)
			}
		}

		srcBucket := cmdParams[0]
		destBucket := cmdParams[1]
		srcDestKeyMapFile := cmdParams[2]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}

		//get bucket zone info
		bucketInfo, gErr := atfuck.GetBucketInfo(&mac, srcBucket)
		if gErr != nil {
			fmt.Println("Get bucket region info error,", gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		//set zone info
		atfuck.SetZone(bucketInfo.Region)

		var batchTasks chan func()
		var initBatchOnce sync.Once

		batchWaitGroup := sync.WaitGroup{}
		initBatchOnce.Do(func() {
			batchTasks = make(chan func(), worker)
			for i := 0; i < worker; i++ {
				go doBatchOperation(batchTasks)
			}
		})

		client := rs.NewMac(&mac)
		fp, err := os.Open(srcDestKeyMapFile)
		if err != nil {
			fmt.Println("Open src dest key map file error")
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		entries := make([]atfuck.MoveEntryPath, 0, BATCH_ALLOW_MAX)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, "\t")
			if len(items) == 1 || len(items) == 2 {
				srcKey := items[0]
				destKey := srcKey
				if len(items) == 2 {
					destKey = items[1]
				}
				if srcKey != "" && destKey != "" {
					entry := atfuck.MoveEntryPath{srcBucket, destBucket, srcKey, destKey}
					entries = append(entries, entry)
				}
			}
			if len(entries) == BATCH_ALLOW_MAX {
				toMoveEntries := make([]atfuck.MoveEntryPath, len(entries))
				copy(toMoveEntries, entries)

				batchWaitGroup.Add(1)
				batchTasks <- func() {
					defer batchWaitGroup.Done()
					batchMove(client, toMoveEntries, overwrite)
				}
				entries = make([]atfuck.MoveEntryPath, 0, BATCH_ALLOW_MAX)
			}
		}
		if len(entries) > 0 {
			toMoveEntries := make([]atfuck.MoveEntryPath, len(entries))
			copy(toMoveEntries, entries)

			batchWaitGroup.Add(1)
			batchTasks <- func() {
				defer batchWaitGroup.Done()
				batchMove(client, toMoveEntries, overwrite)
			}
		}

		batchWaitGroup.Wait()
	} else {
		CmdHelp(cmd)
	}
}

func batchMove(client rs.Client, entries []atfuck.MoveEntryPath, overwrite bool) {
	ret, err := atfuck.BatchMove(client, entries, overwrite)

	if len(ret) > 0 {
		for i, entry := range entries {
			item := ret[i]
			if item.Code != 200 || item.Data.Error != "" {
				logs.Error("Move '%s:%s' => '%s:%s' Failed, Code: %d, Error: %s",
					entry.SrcBucket, entry.SrcKey, entry.DestBucket, entry.DestKey, item.Code, item.Data.Error)
			} else {
				logs.Debug("Move '%s:%s' => '%s:%s' success",
					entry.SrcBucket, entry.SrcKey, entry.DestBucket, entry.DestKey)
			}
		}
	} else {
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Batch move error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Batch move error,", err)
			}
		}
	}
}

func BatchCopy(cmd string, params ...string) {
	var force bool
	var overwrite bool
	var worker int

	flagSet := flag.NewFlagSet("batchcopy", flag.ExitOnError)
	flagSet.BoolVar(&force, "force", false, "force mode")
	flagSet.BoolVar(&overwrite, "overwrite", false, "overwrite mode")
	flagSet.IntVar(&worker, "worker", 1, "worker count")
	flagSet.Parse(params)

	cmdParams := flagSet.Args()
	if len(cmdParams) == 3 {
		if !force {
			//confirm
			rcode := CreateRandString(6)

			rcode2 := ""
			if runtime.GOOS == "windows" {
				fmt.Printf("<DANGER> Input %s to confirm operation: ", rcode)
			} else {
				fmt.Printf("\033[31m<DANGER>\033[0m Input \033[32m%s\033[0m to confirm operation: ", rcode)
			}
			fmt.Scanln(&rcode2)

			if rcode != rcode2 {
				fmt.Println("Task quit!")
				os.Exit(atfuck.STATUS_HALT)
			}
		}

		srcBucket := cmdParams[0]
		destBucket := cmdParams[1]
		srcDestKeyMapFile := cmdParams[2]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}

		//get bucket zone info
		bucketInfo, gErr := atfuck.GetBucketInfo(&mac, srcBucket)
		if gErr != nil {
			fmt.Println("Get bucket region info error,", gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		//set zone info
		atfuck.SetZone(bucketInfo.Region)

		var batchTasks chan func()
		var initBatchOnce sync.Once

		batchWaitGroup := sync.WaitGroup{}
		initBatchOnce.Do(func() {
			batchTasks = make(chan func(), worker)
			for i := 0; i < worker; i++ {
				go doBatchOperation(batchTasks)
			}
		})

		client := rs.NewMac(&mac)
		fp, err := os.Open(srcDestKeyMapFile)
		if err != nil {
			fmt.Println("Open src dest key map file error")
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		entries := make([]atfuck.CopyEntryPath, 0, BATCH_ALLOW_MAX)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, "\t")
			if len(items) == 1 || len(items) == 2 {
				srcKey := items[0]
				destKey := srcKey
				if len(items) == 2 {
					destKey = items[1]
				}
				if srcKey != "" && destKey != "" {
					entry := atfuck.CopyEntryPath{srcBucket, destBucket, srcKey, destKey}
					entries = append(entries, entry)
				}
			}
			if len(entries) == BATCH_ALLOW_MAX {
				toCopyEntries := make([]atfuck.CopyEntryPath, len(entries))
				copy(toCopyEntries, entries)

				batchWaitGroup.Add(1)
				batchTasks <- func() {
					defer batchWaitGroup.Done()
					batchCopy(client, toCopyEntries, overwrite)
				}
				entries = make([]atfuck.CopyEntryPath, 0, BATCH_ALLOW_MAX)
			}
		}
		if len(entries) > 0 {
			toCopyEntries := make([]atfuck.CopyEntryPath, len(entries))
			copy(toCopyEntries, entries)

			batchWaitGroup.Add(1)
			batchTasks <- func() {
				defer batchWaitGroup.Done()
				batchCopy(client, toCopyEntries, overwrite)
			}
		}

		batchWaitGroup.Wait()
	} else {
		CmdHelp(cmd)
	}
}

func batchCopy(client rs.Client, entries []atfuck.CopyEntryPath, overwrite bool) {
	ret, err := atfuck.BatchCopy(client, entries, overwrite)

	if len(ret) > 0 {
		for i, entry := range entries {
			item := ret[i]
			if item.Code != 200 || item.Data.Error != "" {
				logs.Error("Copy '%s:%s' => '%s:%s' Failed, Code: %d, Error: %s",
					entry.SrcBucket, entry.SrcKey, entry.DestBucket, entry.DestKey, item.Code, item.Data.Error)
			} else {
				logs.Debug("Copy '%s:%s' => '%s:%s' success",
					entry.SrcBucket, entry.SrcKey, entry.DestBucket, entry.DestKey)
			}
		}
	} else {
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Printf("Batch copy error, %d %s, xreqid: %s\n", v.Code, v.Err, v.Reqid)
			} else {
				fmt.Println("Batch copy error,", err)
			}
		}
	}
}

func PrivateUrl(cmd string, params ...string) {
	if len(params) == 1 || len(params) == 2 {
		publicUrl := params[0]
		var deadline int64
		if len(params) == 2 {
			if val, err := strconv.ParseInt(params[1], 10, 64); err != nil {
				fmt.Println("Invalid <Deadline>")
				os.Exit(atfuck.STATUS_HALT)
			} else {
				deadline = val
			}
		} else {
			deadline = time.Now().Add(time.Second * 3600).Unix()
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		url := atfuck.PrivateUrl(&mac, publicUrl, deadline)
		fmt.Println(url)
	} else {
		CmdHelp(cmd)
	}
}

func BatchSign(cmd string, params ...string) {
	if len(params) == 1 || len(params) == 2 {
		urlListFile := params[0]
		var deadline int64
		if len(params) == 2 {
			if val, err := strconv.ParseInt(params[1], 10, 64); err != nil {
				fmt.Println("Invalid <Deadline>")
				os.Exit(atfuck.STATUS_HALT)
			} else {
				deadline = val
			}
		} else {
			deadline = time.Now().Add(time.Second * 3600 * 24 * 365).Unix()
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}

		fp, openErr := os.Open(urlListFile)
		if openErr != nil {
			fmt.Println("Open url list file error,", openErr)
			os.Exit(atfuck.STATUS_HALT)
		}
		defer fp.Close()

		bReader := bufio.NewScanner(fp)
		bReader.Split(bufio.ScanLines)
		for bReader.Scan() {
			urlToSign := strings.TrimSpace(bReader.Text())
			if urlToSign == "" {
				continue
			}
			signedUrl := atfuck.PrivateUrl(&mac, urlToSign, deadline)
			fmt.Println(signedUrl)
		}
	} else {
		CmdHelp(cmd)
	}
}

func Saveas(cmd string, params ...string) {
	if len(params) == 3 {
		publicUrl := params[0]
		saveBucket := params[1]
		saveKey := params[2]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
			os.Exit(atfuck.STATUS_ERROR)
		}

		mac := digest.Mac{
			account.AccessKey,
			[]byte(account.SecretKey),
		}
		url, err := atfuck.Saveas(&mac, publicUrl, saveBucket, saveKey)
		if err != nil {
			fmt.Println(err)
			os.Exit(atfuck.STATUS_ERROR)
		} else {
			fmt.Println(url)
		}
	} else {
		CmdHelp(cmd)
	}
}

func M3u8Delete(cmd string, params ...string) {
	if len(params) == 2 {
		bucket := params[0]
		m3u8Key := params[1]

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		m3u8FileList, err := atfuck.M3u8FileList(&mac, bucket, m3u8Key)
		if err != nil {
			fmt.Println(err)
			os.Exit(atfuck.STATUS_ERROR)
		}
		client := rs.NewMac(&mac)
		entryCnt := len(m3u8FileList)
		if entryCnt == 0 {
			fmt.Println("no m3u8 slices found")
			os.Exit(atfuck.STATUS_ERROR)
		}
		if entryCnt <= BATCH_ALLOW_MAX {
			batchDelete(client, m3u8FileList)
		} else {
			batchCnt := entryCnt / BATCH_ALLOW_MAX
			for i := 0; i < batchCnt; i++ {
				end := (i + 1) * BATCH_ALLOW_MAX
				if end > entryCnt {
					end = entryCnt
				}
				entriesToDelete := m3u8FileList[i*BATCH_ALLOW_MAX : end]
				batchDelete(client, entriesToDelete)
			}
		}
	} else {
		CmdHelp(cmd)
	}
}

func M3u8Replace(cmd string, params ...string) {
	if len(params) == 2 || len(params) == 3 {
		bucket := params[0]
		m3u8Key := params[1]
		var newDomain string
		if len(params) == 3 {
			newDomain = strings.TrimRight(params[2], "/")
		}

		account, gErr := atfuck.GetAccount()
		if gErr != nil {
			fmt.Println(gErr)
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

		err := atfuck.M3u8ReplaceDomain(&mac, bucket, m3u8Key, newDomain)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Println("m3u8 replace domain error,", v.Err)
			} else {
				fmt.Println("m3u8 replace domain error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		}
	} else {
		CmdHelp(cmd)
	}
}
