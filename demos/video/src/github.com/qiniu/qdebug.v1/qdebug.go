package qdebug

import (
	"net/http"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/http/rpcutil.v1"
)

var (
	startTime time.Time
)

func init() {
	startTime = time.Now()
}

type service struct {
}

type pingRet struct {
	Uptime       int64  `json:"uptime"`
	Fd           int    `json:"fd"`
	Mem          uint64 `json:"mem"`
	NumGoroutine int    `json:"num_goroutine"`
}

func (s service) GetQdebugPing(env rpcutil.Env) (ret pingRet, err error) {

	ret.Uptime = int64(time.Since(startTime) / time.Second)
	ret.NumGoroutine = runtime.NumGoroutine()

	fdDir, err := os.Open("/proc/self/fd")
	if err == nil {
		fi, err := fdDir.Readdir(-1)
		if err == nil {
			ret.Fd = len(fi)
		}
		fdDir.Close()
	}
	err = nil

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ret.Mem = m.Alloc

	return
}

func (s service) GetDebugPprof_(env *rpcutil.Env) {

	http.DefaultServeMux.ServeHTTP(env.W, env.Req)
}

func QDebug(addr string) (err error) {

	router := restrpc.Router{}
	serveOpt := httputil.ServeOption{
		StopTimeout: time.Second,
		KillTimeout: time.Second,
	}

	err = httputil.ListenAndServe(addr, router.Register(service{}), serveOpt)
	return
}
