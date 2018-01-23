package gobrpc

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"reflect"
	"strconv"

	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/rpcutil.v1"
)

type Env rpcutil.Env

/* ---------------------------------------------------------------------------

func (rcvr *XXXX) GobrpcYYYY(req ZZZZ, env ENV) (err error)
func (rcvr *XXXX) GobrpcYYYY(req ZZZZ, env ENV) (ret RRRR, err error)

func (rcvr *XXXX) GobrpcYYYY(req ZZZZ) (err error)
func (rcvr *XXXX) GobrpcYYYY(req ZZZZ) (ret RRRR, err error)

func (rcvr *XXXX) GobrpcYYYY(env ENV) (err error)
func (rcvr *XXXX) GobrpcYYYY(env ENV) (ret RRRR, err error)

func (rcvr *XXXX) GobrpcYYYY() (err error)
func (rcvr *XXXX) GobrpcYYYY() (ret RRRR, err error)

// -------------------------------------------------------------------------*/

type rpcError interface {
	RpcError() (code, errno int, key, err string)
}

func Error(w http.ResponseWriter, err error) {

	h := w.Header()
	h.Set("Content-Length", "0")
	h.Set("Content-Type", "application/gob")

	if e, ok := err.(rpcError); ok {
		code, errno, _, emsg := e.RpcError()
		h.Set("X-Err", emsg)
		h.Set("X-Errno", strconv.Itoa(errno))
		w.WriteHeader(code)
	} else {
		h.Set("X-Err", err.Error())
		w.WriteHeader(599)
	}
}

func Reply(w http.ResponseWriter, code int, data interface{}) {

	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(data)
	if err != nil {
		Error(w, err)
		return
	}

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(b.Len()))
	h.Set("Content-Type", "application/gob")
	w.WriteHeader(code)
	w.Write(b.Bytes())
}

func ReplyWithCode(w http.ResponseWriter, code int) {

	h := w.Header()
	h.Set("Content-Length", "0")
	h.Set("Content-Type", "application/gob")
	w.WriteHeader(code)
}

var gobRepl = &rpcutil.Replier{
	Reply:         Reply,
	ReplyWithCode: ReplyWithCode,
	Error:         Error,
}

// ---------------------------------------------------------------------------

func ParseGobReq(v reflect.Value, req *http.Request) error {
	return parseGobReq(v, req)
}

func parseGobReq(v reflect.Value, req *http.Request) error {

	if req.ContentLength == 0 {
		return nil
	}
	return gob.NewDecoder(req.Body).Decode(v.Interface())
}

var Factory = hfac.HandlerFactory{
	{"Gobrpc", rpcutil.HandlerCreator{
		ParseReq:     parseGobReq,
		ReqMayNotPtr: true,
		Repl:         gobRepl,
	}.New},
	{"Do", hfac.NewHandler},
}

// ---------------------------------------------------------------------------
