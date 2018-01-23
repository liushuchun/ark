package rpc_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v2"
	"github.com/stretchr/testify/assert"
)

var userAgentTst string

func foo(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	httputil.Reply(w, 200, map[string]interface{}{
		"info":         "Call method foo",
		"url":          req.RequestURI,
		"query":        req.Form,
		"content-type": req.Header.Get("Content-Type"),
	})
}

func agent(w http.ResponseWriter, req *http.Request) {

	userAgentTst = req.Header.Get("User-Agent")
}

type Object struct {
}

func (p *Object) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req2, _ := ioutil.ReadAll(req.Body)
	httputil.Reply(w, 200, map[string]interface{}{
		"info": "Call method object",
		"req":  string(req2),
	})
}

var done = make(chan bool)

func server(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", foo)
	mux.Handle("object", new(Object))
	return httptest.NewServer(mux)
}

func TestCall(t *testing.T) {
	s := server(t)
	defer s.Close()

	//param := "http:**localhost:8888*abc:def,g;+&$=foo*~!*~!"
	r := map[string]interface{}{}
	c := rpc.DefaultClient
	c.Call(nil, &r, "GET", s.URL+"/foo")
	assert.Equal(t, r, map[string]interface{}{"info": "Call method foo", "query": map[string]interface{}{}, "url": "/foo", "content-type": "application/x-www-form-urlencoded"})

	c.CallWithForm(nil, &r, "GET", s.URL+"/foo", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?a=1")

	c.CallWithForm(nil, &r, "GET", s.URL+"/foo?b=2", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?b=2&a=1")

	c.CallWithForm(nil, &r, "GET", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?&a=1")

	resp, err := c.DoRequestWithForm(nil, "GET", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, err, nil)
	err = rpc.CallRet(nil, &r, resp)
	assert.Equal(t, err, nil)
	assert.Equal(t, r["content-type"], "")

	resp, err = c.DoRequestWithForm(nil, "POST", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, err, nil)
	err = rpc.CallRet(nil, &r, resp)
	assert.Equal(t, err, nil)
	assert.Equal(t, r["content-type"], "application/x-www-form-urlencoded")

	resp, err = c.DoRequestWithForm(nil, "DELETE", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, err, nil)
	err = rpc.CallRet(nil, &r, resp)
	assert.Equal(t, err, nil)
	assert.Equal(t, r["content-type"], "")
}

func TestDo(t *testing.T) {

	svr := httptest.NewServer(http.HandlerFunc(agent))
	defer svr.Close()

	svrUrl := svr.URL
	c := rpc.DefaultClient
	{
		req, _ := http.NewRequest("GET", svrUrl+"/agent", nil)
		c.Do(nil, req)
		assert.Equal(t, userAgentTst, "Golang qiniu/rpc package")
	}
	{
		req, _ := http.NewRequest("GET", svrUrl+"/agent", nil)
		req.Header.Set("User-Agent", "tst")
		c.Do(nil, req)
		assert.Equal(t, userAgentTst, "tst")
	}
}
