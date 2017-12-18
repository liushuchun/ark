package main

import (
	"atfuck"
	"cli"
	"flag"
	"fmt"
	"os"
	"os/user"
	"qiniu/rpc"
	"runtime"

	"github.com/astaxie/beego/logs"
)

var supportedCmds = map[string]cli.CliFunc{
	"acc":        cli.Account,
	"d":          cli.QiniuDownload,
	"qetag":      cli.Qetag,
	"unzip":      cli.Unzip,
	"privateurl": cli.PrivateUrl,
	"saveas":     cli.Saveas,
}

func main() {
	//set cpu count
	runtime.GOMAXPROCS(runtime.NumCPU())
	//set atfuck user agent
	rpc.UserAgent = cli.UserAgent()

	//parse command
	logs.SetLevel(logs.LevelInformational)
	logs.SetLogger(logs.AdapterConsole)

	if len(os.Args) <= 1 {
		fmt.Println("you should input the params")
		os.Exit(atfuck.STATUS_HALT)
	}

	//global options
	var debugMode bool
	var helpMode bool
	var versionMode bool
	var multiUserMode bool
	var unzip bool
	flag.BoolVar(&debugMode, "d", false, "debug mode")
	flag.BoolVar(&multiUserMode, "m", false, "multi user mode")
	flag.BoolVar(&helpMode, "h", false, "show help")
	flag.BoolVar(&versionMode, "v", false, "show version")
	flag.BoolVar(&unzip, "unzip", false, "unzip the file to")
	flag.Parse()

	if helpMode {
		cli.Help("help")
		return
	}

	if versionMode {
		cli.Version()
		return
	}

	//set log level
	if debugMode {
		logs.SetLevel(logs.LevelDebug)
	}

	//set atfuck root path
	if multiUserMode {
		logs.Debug("Entering multiple user mode")
		pwd, gErr := os.Getwd()
		if gErr != nil {
			fmt.Println("Error: get current work dir error,", gErr)
			os.Exit(atfuck.STATUS_HALT)
		}
		atfuck.QShellRootPath = pwd
	} else {
		logs.Debug("Entering single user mode")
		curUser, gErr := user.Current()
		if gErr != nil {
			fmt.Println("Error: get current user error,", gErr)
			os.Exit(atfuck.STATUS_HALT)
		}
		atfuck.QShellRootPath = curUser.HomeDir
	}

	//set cmd and params
	args := flag.Args()
	cmd := args[0]
	params := args[1:]

	if cliFunc, ok := supportedCmds[cmd]; ok {
		cliFunc(cmd, params...)
	} else {
		fmt.Printf("Error: unknown cmd `%s`\n", cmd)
		os.Exit(atfuck.STATUS_HALT)
	}
}
