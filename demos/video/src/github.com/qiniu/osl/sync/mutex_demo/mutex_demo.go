package main

import (
	"fmt"
	"github.com/qiniu/osl/sync"
	"time"
)

func main() {

	fmt.Println("Begin... 1")

	f, err := sync.CreateMutex("qiniu.test.pid")
	if err != nil {
		fmt.Println("CreateMutex failed:", err)
		return
	}

	fmt.Scanln()

	f.Close()

	fmt.Println("Begin... 2")

	time.Sleep(time.Hour * 2400 * 365)
}
