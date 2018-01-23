package jsonrpc

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/rpcutil.v1"
)

type Env rpcutil.Env

/* ---------------------------------------------------------------------------

func (rcvr *XXXX) RpcYYYY(req ZZZZ, env ENV) (err error)
func (rcvr *XXXX) RpcYYYY(req ZZZZ, env ENV) (ret RRRR, err error)

func (rcvr *XXXX) RpcYYYY(req ZZZZ) (err error)
func (rcvr *XXXX) RpcYYYY(req ZZZZ) (ret RRRR, err error)

func (rcvr *XXXX) RpcYYYY(env ENV) (err error)
func (rcvr *XXXX) RpcYYYY(env ENV) (ret RRRR, err error)

func (rcvr *XXXX) RpcYYYY() (err error)
func (rcvr *XXXX) RpcYYYY() (ret RRRR, err error)

// -------------------------------------------------------------------------*/

func ParseJsonReq(v reflect.Value, req *http.Request) error {
	return parseJsonReq(v, req)
}

func parseJsonReq(v reflect.Value, req *http.Request) error {

	if req.ContentLength == 0 {
		return nil
	}
	return json.NewDecoder(req.Body).Decode(v.Interface())
}

var Factory = hfac.HandlerFactory{
	{"Rpc", rpcutil.HandlerCreator{
		ParseReq:     parseJsonReq,
		ReqMayNotPtr: true,
	}.New},
	{"Do", hfac.NewHandler},
}

// ---------------------------------------------------------------------------
