package main

import (
	"github.com/qiniu/http/unidi/unidisvr"
	"github.com/qiniu/log.v1"
	"io"
	"net/http"
)

// ---------------------------------------------------------------------------

func echo(w http.ResponseWriter, req *http.Request) {

	name := req.FormValue("name")
	log.Info("echo:", name)

	io.WriteString(w, name)
}

// ---------------------------------------------------------------------------

func main() {

	http.HandleFunc("/echo", echo)
	unidisvr.ListenAndServe("unidisvr1", "http://localhost:9630/unidi", nil)
}

// ---------------------------------------------------------------------------
