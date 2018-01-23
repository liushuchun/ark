package main

import (
	"fmt"
	"github.com/qiniu/http/unidi/unidicli"
	"github.com/qiniu/log.v1"
	"io"
	"net/http"
	"os"
)

// ---------------------------------------------------------------------------

func main() {

	tr := unidicli.NewTransport(1e9)
	unidicli.RegisterProtocol("unidi", tr)
	http.Handle("/unidi", tr)
	go func() {
		err := http.ListenAndServe(":9630", nil)
		log.Fatal("ListenAndServe:", err)
	}()

	for {
		var line string
		fmt.Scanf("%s", &line)
		if line == "q" {
			break
		}
		resp, err := http.Get("unidi://unidisvr1/echo?name=" + line)
		if err != nil {
			log.Warn("http.Get failed:", err)
			continue
		}
		io.Copy(os.Stdout, resp.Body)
		fmt.Println()
		resp.Body.Close()
	}
}

// ---------------------------------------------------------------------------
