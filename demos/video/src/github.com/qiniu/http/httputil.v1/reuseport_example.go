// +build ignore

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/qiniu/http/httputil.v1"
)

func handler(w http.ResponseWriter, r *http.Request) {

	// curl 'http://localhost:8080/?duration=20s'

	fmt.Println("req")
	duration, err := time.ParseDuration(r.FormValue("duration"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Fprintf(w, "going to sleep %s with pid %d\n", duration, os.Getpid())
	w.(http.Flusher).Flush()
	time.Sleep(duration)
	fmt.Fprintf(w, "slept %s with pid %d\n", duration, os.Getpid())
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	var addr = "127.0.0.1:8080"
	var stopTimeout = 10 * time.Second
	var killTimeout = 10 * time.Second

	flag.StringVar(&addr, "addr", addr, "http address")
	flag.DurationVar(&stopTimeout, "stop-timeout", stopTimeout, "stop timeout")
	flag.DurationVar(&killTimeout, "kill-timeout", killTimeout, "kill timeout")
	flag.Parse()

	opt := httputil.ServeOption{
		StopFunc: func() error {
			fmt.Println("stopping...")
			return nil
		},
		ForceReusePort: true,
		StopTimeout:    stopTimeout,
		KillTimeout:    killTimeout,
	}

	fmt.Printf("serve @%s\n", addr)
	err := httputil.ListenAndServe(addr, mux, opt)
	if err != nil {
		fmt.Println("ListenAndServe failed:", err)
	}
}
