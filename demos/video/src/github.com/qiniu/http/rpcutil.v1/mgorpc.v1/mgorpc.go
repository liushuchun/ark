package mgorpc_v1

import (
	"strings"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/servestk.v1"
)

const ErrMgoNoRechableServers = 560

var DefaultHandleErr = func(err error) error {
	if strings.Contains(err.Error(), "no reachable servers") {
		return httputil.NewError(ErrMgoNoRechableServers, err.Error())
	}
	return err
}

var DefaultHandlePanic = func(v interface{}) error {
	if s, ok := v.(string); ok {
		if strings.Contains(s, "[MGO2_COPY_SESSION_FAILED] servers failed") {
			return httputil.NewError(ErrMgoNoRechableServers, "no reachable servers")
		}
	}
	return httputil.NewError(597, "internal failed")
}

func init() {
	rpcutil.HandleErr = DefaultHandleErr
	servestk.HandlePanic = DefaultHandlePanic
}

// ---------------------------------------------------------------------------
