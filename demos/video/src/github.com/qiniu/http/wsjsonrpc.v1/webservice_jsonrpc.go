package wsjsonrpc

import (
	"net/http"
	"reflect"

	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/jsonrpc.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/wsrpc.v1"
)

type Env rpcutil.Env

var Factory = hfac.HandlerFactory{
	{"Wsprpc", rpcutil.HandlerCreator{
		ParseReq: func(v reflect.Value, req *http.Request) error {
			err := formutil.ParseForm(v.Interface(), req, false)
			if err != nil {
				return err
			}
			return jsonrpc.ParseJsonReq(v, req)
		},
		ReqMayNotPtr: true,
		PostOnly:     true,
	}.New},
	{"Cmdprpc", rpcutil.HandlerCreator{
		ParseReq: func(v reflect.Value, req *http.Request) error {
			err := wsrpc.ParseCmdReq(v, req)
			if err != nil {
				return err
			}
			return jsonrpc.ParseJsonReq(v, req)
		},
		ReqMayNotPtr: true,
		PostOnly:     true,
	}.New},
	{"Do", hfac.NewHandler},
}
