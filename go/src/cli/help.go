package cli

import (
	"fmt"
	"os"
	"runtime"
)

var version = "v2.0.9"

var optionDocs = map[string]string{
	"-f": "Force batch operations",
	"-d": "Show debug message",
	"-v": "Show version",
	"-h": "Show help",
}

var cmds = []string{
	"account",
	"zone",
	"dircache",
	"listbucket",
	"prefop",
	"fput",
	"rput",
	"qupload",
	"qupload2",
	"qdownload",
	"stat",
	"delete",
	"move",
	"copy",
	"chgm",
	"fetch",
	"sync",
	"prefetch",
	"batchstat",
	"batchdelete",
	"batchchgm",
	"batchcopy",
	"batchmove",
	"batchrename",
	"batchsign",
	"privateurl",
	"saveas",
	"reqid",
	"buckets",
	"domains",
	"qetag",
	"m3u8delete",
	"m3u8replace",
	"cdnrefresh",
	"cdnprefetch",
	"b64encode",
	"b64decode",
	"urlencode",
	"urldecode",
	"ts2d",
	"tms2d",
	"tns2d",
	"d2ts",
	"ip",
	"unzip",
	"alilistbucket",
}
var cmdDocs = map[string][]string{
	"account":       {"atfuck account [<AccessKey> <SecretKey>] [<Zone>]", "Get/Set AccessKey and SecretKey and Zone"},
	"zone":          {"atfuck zone [<Zone>]", "Switch the zone, [nb, bc, hn, na0]"},
	"dircache":      {"atfuck dircache <DirCacheRootPath> <DirCacheResultFile>", "Cache the directory structure of a file path"},
	"listbucket":    {"atfuck listbucket [-marker <ListMarker>] <Bucket> [<Prefix>] <ListBucketResultFile>", "List all the files in the bucket by prefix"},
	"alilistbucket": {"atfuck alilistbucket <DataCenter> <Bucket> <AccessKeyId> <AccesskeySecret> [Prefix] <ListBucketResultFile>", "List all the file in the bucket of aliyun oss by prefix"},
	"prefop":        {"atfuck prefop <PersistentId>", "Query the pfop status"},
	"fput":          {"atfuck fput <Bucket> <Key> <LocalFile> [<Overwrite>] [<MimeType>] [<UpHost>] [<FileType>]", "Form upload a local file"},
	"rput":          {"atfuck rput <Bucket> <Key> <LocalFile> [<Overwrite>] [<MimeType>] [<UpHost>] [<FileType>]", "Resumable upload a local file"},
	"qupload":       {"atfuck qupload [<ThreadCount>] <LocalUploadConfig>", "Batch upload files to the qiniu bucket"},
	"qupload2":      {"atfuck qupload2 [options]", "Batch upload files to the qiniu bucket"},
	"qdownload":     {"atfuck qdownload [<ThreadCount>] <LocalDownloadConfig>", "Batch download files from the qiniu bucket"},
	"stat":          {"atfuck stat <Bucket> <Key>", "Get the basic info of a remote file"},
	"delete":        {"atfuck delete <Bucket> <Key>", "Delete a remote file in the bucket"},
	"move":          {"atfuck move [-overwrite] <SrcBucket> <SrcKey> <DestBucket> [<DestKey>]", "Move/Rename a file and save in bucket"},
	"copy":          {"atfuck copy [-overwrite] <SrcBucket> <SrcKey> <DestBucket> [<DestKey>]", "Make a copy of a file and save in bucket"},
	"chgm":          {"atfuck chgm <Bucket> <Key> <NewMimeType>", "Change the mimeType of a file"},
	"sync":          {"atfuck sync <SrcResUrl> <Bucket> <Key> [<UpHostIp>]", "Sync big file to qiniu bucket"},
	"fetch":         {"atfuck fetch <RemoteResourceUrl> <Bucket> [<Key>]", "Fetch a remote resource by url and save in bucket"},
	"prefetch":      {"atfuck prefetch <Bucket> <Key>", "Fetch and update the file in bucket using mirror storage"},
	"batchstat":     {"atfuck batchstat <Bucket> <KeyListFile>", "Batch stat files in bucket"},
	"batchdelete":   {"atfuck batchdelete [-force] <Bucket> <KeyListFile>", "Batch delete files in bucket"},
	"batchchgm":     {"atfuck batchchgm [-force] <Bucket> <KeyMimeMapFile>", "Batch chgm files in bucket"},
	"batchcopy":     {"atfuck batchcopy [-force] [-overwrite] <SrcBucket> <DestBucket> <SrcDestKeyMapFile>", "Batch copy files from bucket to bucket"},
	"batchmove":     {"atfuck batchmove [-force] [-overwrite] <SrcBucket> <DestBucket> <SrcDestKeyMapFile>", "Batch move files from bucket to bucket"},
	"batchrename":   {"atfuck batchrename [-force] [-overwrite] <Bucket> <OldNewKeyMapFile>", "Batch rename files in the bucket"},
	"batchsign":     {"atfuck batchsign <UrlListFile> [<Deadline>]", "Batch create the private url from the public url list file"},
	"b64encode":     {"atfuck b64encode [<UrlSafe>] <DataToEncode>", "Base64 Encode"},
	"b64decode":     {"atfuck b64decode [<UrlSafe>] <DataToDecode>", "Base64 Decode"},
	"urlencode":     {"atfuck urlencode <DataToEncode>", "Url encode"},
	"urldecode":     {"atfuck urldecode <DataToDecode>", "Url decode"},
	"ts2d":          {"atfuck ts2d <TimestampInSeconds>", "Convert timestamp in seconds to a date (TZ: Local)"},
	"tms2d":         {"atfuck tms2d <TimestampInMilliSeconds>", "Convert timestamp in milli-seconds to a date (TZ: Local)"},
	"tns2d":         {"atfuck tns2d <TimestampIn100NanoSeconds>", "Convert timestamp in 100 nano-seconds to a date (TZ: Local)"},
	"d2ts":          {"atfuck d2ts <SecondsToNow>", "Create a timestamp in seconds using seconds to now"},
	"ip":            {"atfuck ip <Ip1> [<Ip2> [<Ip3> ...]]]", "Query the ip information"},
	"qetag":         {"atfuck qetag <LocalFilePath>", "Calculate the hash of local file using the algorithm of qiniu qetag"},
	"unzip":         {"atfuck unzip <QiniuZipFilePath> [<UnzipToDir>]", "Unzip the archive file created by the qiniu mkzip API"},
	"privateurl":    {"atfuck privateurl <PublicUrl> [<Deadline>]", "Create private resource access url"},
	"saveas":        {"atfuck saveas <PublicUrlWithFop> <SaveBucket> <SaveKey>", "Create a resource access url with fop and saveas"},
	"reqid":         {"atfuck reqid <ReqIdToDecode>", "Decode a qiniu reqid"},
	"m3u8delete":    {"atfuck m3u8delete <Bucket> <M3u8Key>", "Delete m3u8 playlist and the slices it references"},
	"m3u8replace":   {"atfuck m3u8replace <Bucket> <M3u8Key> [<NewDomain>]", "Replace m3u8 domain in the playlist"},
	"buckets":       {"atfuck buckets", "Get all buckets of the account"},
	"domains":       {"atfuck domains <Bucket>", "Get all domains of the bucket"},
	"cdnrefresh":    {"atfuck cdnrefresh <UrlListFile>", "Batch refresh the cdn cache by the url list file"},
	"cdnprefetch":   {"atfuck cdnprefetch <UrlListFile>", "Batch prefetch the urls in the url list file"},
}

func Version() {
	fmt.Printf("QShell/%s (%s; %s; %s)\n", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func Help(cmd string, params ...string) {
	if len(params) == 0 {
		fmt.Println(CmdList())
	} else {
		CmdHelps(params...)
	}
}

func CmdList() string {
	helpAll := fmt.Sprintf("QShell %s\r\n\r\n", version)
	helpAll += "Options:\r\n"
	for k, v := range optionDocs {
		helpAll += fmt.Sprintf("\t%-20s%-20s\r\n", k, v)
	}
	helpAll += "\r\n"
	helpAll += "Commands:\r\n"
	for _, cmd := range cmds {
		if help, ok := cmdDocs[cmd]; ok {
			cmdDesc := help[1]
			helpAll += fmt.Sprintf("\t%-20s%-20s\r\n", cmd, cmdDesc)
		}
	}
	return helpAll
}

func CmdHelps(cmds ...string) {
	defer os.Exit(1)
	if len(cmds) == 0 {
		fmt.Println(CmdList())
	} else {
		for _, cmd := range cmds {
			CmdHelp(cmd)
		}
	}
}

func CmdHelp(cmd string) {
	docStr := fmt.Sprintf("Unknow cmd `%s`", cmd)
	if cmdDoc, ok := cmdDocs[cmd]; ok {
		docStr = fmt.Sprintf("Usage: %s\r\n  %s\r\n", cmdDoc[0], cmdDoc[1])
	}
	fmt.Println(docStr)
}

func UserAgent() string {
	return fmt.Sprintf("QShell/%s (%s; %s; %s)", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
