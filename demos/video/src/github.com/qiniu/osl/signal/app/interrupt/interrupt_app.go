package main

import (
	"github.com/qiniu/osl/signal"
	"os"
)

func main() {

	signal.WaitForInterrupt(func() {
		os.Exit(0)
	})
}
