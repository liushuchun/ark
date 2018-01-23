package wsrpc

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/qiniu/http/flag.v1"
	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/rpcutil.v1"
)

type Env rpcutil.Env

/* ---------------------------------------------------------------------------

allow GET/POST:

func (rcvr *XXXX) WsYYYY(req *ZZZZ, env ENV) (err error)
func (rcvr *XXXX) WsYYYY(req *ZZZZ, env ENV) (ret RRRR, err error)
func (rcvr *XXXX) WsYYYY(req *ZZZZ, env ENV)

func (rcvr *XXXX) WsYYYY(req *ZZZZ) (err error)
func (rcvr *XXXX) WsYYYY(req *ZZZZ) (ret RRRR, err error)

func (rcvr *XXXX) WsYYYY(env ENV) (err error)
func (rcvr *XXXX) WsYYYY(env ENV) (ret RRRR, err error)

func (rcvr *XXXX) WsYYYY() (err error)
func (rcvr *XXXX) WsYYYY() (ret RRRR, err error)

allow POST only:

func (rcvr *XXXX) WspYYYY(req *ZZZZ, env ENV) (err error)
func (rcvr *XXXX) WspYYYY(req *ZZZZ, env ENV) (ret RRRR, err error)
func (rcvr *XXXX) WspYYYY(req *ZZZZ, env ENV)

func (rcvr *XXXX) WspYYYY(req *ZZZZ) (err error)
func (rcvr *XXXX) WspYYYY(req *ZZZZ) (ret RRRR, err error)

func (rcvr *XXXX) WspYYYY(env ENV) (err error)
func (rcvr *XXXX) WspYYYY(env ENV) (ret RRRR, err error)

func (rcvr *XXXX) WspYYYY() (err error)
func (rcvr *XXXX) WspYYYY() (ret RRRR, err error)

// -------------------------------------------------------------------------*/

func ParseCmdReq(v reflect.Value, req *http.Request) error {
	return parseCmdReq(v, req)
}

func parseCmdReq(v reflect.Value, req *http.Request) error {

	return flag.ParseValue(v, strings.Split(req.URL.Path[1:], "/"), "flag")
}

var Factory = hfac.HandlerFactory{
	{"Wsp", rpcutil.HandlerCreator{
		ParseReq: func(v reflect.Value, req *http.Request) error {
			return formutil.ParseForm(v.Interface(), req, true)
		},
		PostOnly: true,
	}.New},
	{"Cmdp", rpcutil.HandlerCreator{
		ParseReq: parseCmdReq,
		PostOnly: true,
	}.New},
	{"Ws", rpcutil.HandlerCreator{
		ParseReq: func(v reflect.Value, req *http.Request) error {
			return formutil.ParseForm(v.Interface(), req, false)
		},
	}.New},
	{"Cmd", rpcutil.HandlerCreator{
		ParseReq: parseCmdReq,
	}.New},
	{"Do", hfac.NewHandler},
}

// ---------------------------------------------------------------------------
