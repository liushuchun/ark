package bsonrpc

import (
	"github.com/qiniu/http/flag.v1"
	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"gopkg.in/mgo.v2/bson"

	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type Env rpcutil.Env

/* ---------------------------------------------------------------------------

Bbrpc: req = Bson, ret = Bson
Wbrpc: req = Form, ret = Bson

// -------------------------------------------------------------------------*/

func Reply(w http.ResponseWriter, code int, data interface{}) {

	msg, err := bson.Marshal(data)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", "application/bson")
	w.WriteHeader(code)
	w.Write(msg)
}

func ReplyWithCode(w http.ResponseWriter, code int) {

	if code < 400 {
		h := w.Header()
		h.Set("Content-Length", "0")
		h.Set("Content-Type", "application/bson")
		w.WriteHeader(code)
	} else {
		err := http.StatusText(code)
		if err == "" {
			err = "E" + strconv.Itoa(code)
		}
		httputil.ReplyErr(w, code, err)
	}
}

var bsonRepl = &rpcutil.Replier{
	Reply:         Reply,
	ReplyWithCode: ReplyWithCode,
	Error:         httputil.Error,
}

// ---------------------------------------------------------------------------

func parseBsonReq(v reflect.Value, req *http.Request) error {

	if req.ContentLength == 0 {
		return nil
	}
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return bson.Unmarshal(b, v.Interface())
}

func parseCmdReq(v reflect.Value, req *http.Request) error {

	return flag.ParseValue(v, strings.Split(req.URL.Path[1:], "/"), "flag")
}

var Factory = hfac.HandlerFactory{
	{"Bbrpc", rpcutil.HandlerCreator{
		ParseReq: parseBsonReq,
		Repl:     bsonRepl,
	}.New},
	{"Wbrpc", rpcutil.HandlerCreator{
		ParseReq: func(v reflect.Value, req *http.Request) error {
			return formutil.ParseForm(v.Interface(), req, true)
		},
		Repl: bsonRepl,
	}.New},
	{"Cmdbrpc", rpcutil.HandlerCreator{
		ParseReq: parseCmdReq,
		Repl:     bsonRepl,
	}.New},
	{"Cmdpbrpc", rpcutil.HandlerCreator{
		ParseReq: parseCmdReq,
		Repl:     bsonRepl,
		PostOnly: true,
	}.New},
	{"Do", hfac.NewHandler},
}

// ---------------------------------------------------------------------------
