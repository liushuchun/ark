// +build ignore

// run cmd:
// go run -ldflags "-X github.com/qiniu/version.version as" main.go -version
// go run -ldflags "-X github.com/qiniu/version.version as" main.go
package main

import (
	"fmt"
	"github.com/qiniu/version"
)

func main() {
	fmt.Println("from main:", version.Version())
}
