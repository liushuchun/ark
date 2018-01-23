package xhrpc

import (
	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/rpcutil.v1"

	"net/http"
	"net/url"
	"reflect"
)

type Env rpcutil.Env

/* ---------------------------------------------------------------------------

allow GET/POST:

func (rcvr *XXXX) XhYYYY(req *ZZZZ, env ENV) (err error)
func (rcvr *XXXX) XhYYYY(req *ZZZZ, env ENV) (ret RRRR, err error)
func (rcvr *XXXX) XhYYYY(req *ZZZZ, env ENV)

// -------------------------------------------------------------------------*/

var Factory = hfac.HandlerFactory{
	{"Xh", rpcutil.HandlerCreator{
		ParseReq: func(v reflect.Value, req *http.Request) error {
			return formutil.ParseEx(v.Interface(), url.Values(req.Header), "xh")
		},
	}.New},
	{"Do", hfac.NewHandler},
}

// ---------------------------------------------------------------------------

