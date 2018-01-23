package mockhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/rpc.v1"
)

// --------------------------------------------------------------------

func reply(w http.ResponseWriter, code int, data interface{}) {

	msg, _ := json.Marshal(data)
	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
}

// --------------------------------------------------------------------

type FooRet struct {
	A int    `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
}

type HandleRet map[string]string

type FooServer struct{}

func (p *FooServer) foo(w http.ResponseWriter, req *http.Request) {
	reply(w, 200, &FooRet{1, req.Host, req.URL.Path})
}

func (p *FooServer) handle(w http.ResponseWriter, req *http.Request) {
	reply(w, 200, HandleRet{"foo": "1", "bar": "2"})
}

func (p *FooServer) postDump(w http.ResponseWriter, req *http.Request) {
	req.Body.Close()
	io.Copy(w, req.Body)
}

func (p *FooServer) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) { p.foo(w, req) })
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { p.handle(w, req) })
	mux.HandleFunc("/dump", func(w http.ResponseWriter, req *http.Request) { p.postDump(w, req) })
}

func (p *FooServer) AnotherRegisterHandlers(mux *http.ServeMux, path string) {
	fmt.Println("AnotherRegisterHandlers")
	mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) { p.foo(w, req) })
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { p.handle(w, req) })
}

func TestBasic(t *testing.T) {

	server := new(FooServer)
	Bind("foo.com", server)
	BindEx("bar.com", server, "AnotherRegisterHandlers", "/bar")

	c := rpc.DefaultClient
	{
		var foo FooRet
		err := c.Call(nil, &foo, "http://foo.com/foo")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if foo.A != 1 || foo.B != "foo.com" || foo.C != "/foo" {
			t.Fatal("call foo: invalid ret")
		}
		fmt.Println(foo)
	}
	{
		var ret map[string]string
		err := c.Call(nil, &ret, "http://foo.com/bar")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if ret["foo"] != "1" || ret["bar"] != "2" {
			t.Fatal("call bar: invalid ret")
		}
		fmt.Println(ret)
	}
	{
		var ret map[string]string
		err := c.Call(nil, &ret, "http://bar.com/foo")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if ret["foo"] != "1" || ret["bar"] != "2" {
			t.Fatal("call bar: invalid ret")
		}
		fmt.Println(ret)
	}
	{
		var foo FooRet
		err := c.Call(nil, &foo, "http://bar.com/bar")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if foo.A != 1 || foo.B != "bar.com" || foo.C != "/bar" {
			t.Fatal("call foo: invalid ret")
		}
		fmt.Println(foo)
	}
	{
		resp, err := c.Post("http://foo.com/dump", "", nil)
		if err != nil {
			t.Fatal("post foo failed:", err)
		}
		resp.Body.Close()
		resp, err = c.Post("http://foo.com/dump", "", strings.NewReader("abc"))
		if err != nil {
			t.Fatal("post foo failed:", err)
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("ioutil.ReadAll:", err)
		}
		if len(b) != 0 {
			t.Fatal("body should be empty:", string(b))
		}
	}
}

// --------------------------------------------------------------------

type WebrouteServer struct{}

func (r *WebrouteServer) DoFoo(w http.ResponseWriter, req *http.Request) {
	reply(w, 200, &FooRet{2, req.Host, req.URL.Path})
}

func (p *WebrouteServer) Do_(w http.ResponseWriter, req *http.Request) {
	reply(w, 200, HandleRet{"foo": "3", "bar": "4"})
}

func TestWebroute(t *testing.T) {

	server := new(WebrouteServer)

	router := &webroute.Router{}
	Bind("web.com", router.Register(server))

	c := Client
	{
		var foo FooRet
		err := c.Call(nil, &foo, "http://web.com/foo")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if foo.A != 2 || foo.B != "web.com" || foo.C != "/foo" {
			t.Fatal("call foo: invalid ret")
		}
		fmt.Println(foo)
	}
	{
		var ret map[string]string
		err := c.Call(nil, &ret, "http://web.com/bar")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if ret["foo"] != "3" || ret["bar"] != "4" {
			t.Fatal("call bar: invalid ret")
		}
		fmt.Println(ret)
	}
}

// --------------------------------------------------------------------
