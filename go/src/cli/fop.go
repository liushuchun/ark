package cli

import (
	"atfuck"
	"fmt"
	"os"
	"qiniu/rpc"
)

func Prefop(cmd string, params ...string) {
	if len(params) == 1 {
		persistentId := params[0]
		fopRet := atfuck.FopRet{}
		err := atfuck.Prefop(persistentId, &fopRet)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Println("Prefop error,", v.Code, v.Err)
			} else {
				fmt.Println("Prefop error,", err)
			}
			os.Exit(atfuck.STATUS_ERROR)
		} else {
			fmt.Println(fopRet.String())
		}
	} else {
		CmdHelp(cmd)
	}
}
