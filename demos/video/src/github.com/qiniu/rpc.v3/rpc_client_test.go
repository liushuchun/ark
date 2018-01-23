package rpc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.google.com/p/go.net/context"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
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
	c := DefaultClient
	c.Call(nil, &r, "GET", s.URL+"/foo")
	assert.Equal(t, r, map[string]interface{}{"info": "Call method foo", "query": map[string]interface{}{}, "url": "/foo", "content-type": ""})
	c.Call(nil, &r, "POST", s.URL+"/foo")
	assert.Equal(t, r, map[string]interface{}{"info": "Call method foo", "query": map[string]interface{}{}, "url": "/foo", "content-type": "application/x-www-form-urlencoded"})

	c.CallWithForm(nil, &r, "GET", s.URL+"/foo", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?a=1")

	err := c.CallWithForm(nil, &r, "GET", s.URL+"/foo?b=2", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?b=2&a=1")

	c.CallWithForm(nil, &r, "GET", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?&a=1")

	resp, err := c.DoRequestWithForm(nil, "GET", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, err, nil)
	err = CallRet(nil, &r, resp)
	assert.Equal(t, err, nil)
	assert.Equal(t, r["content-type"], "")

	resp, err = c.DoRequestWithForm(nil, "POST", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Nil(t, err)
	err = CallRet(nil, &r, resp)
	assert.Equal(t, err, nil)
	assert.Equal(t, r["content-type"], "application/x-www-form-urlencoded")

	resp, err = c.DoRequestWithForm(nil, "DELETE", s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, err, nil)
	err = CallRet(nil, &r, resp)
	assert.Equal(t, err, nil)
	assert.Equal(t, r["content-type"], "")
}

func TestDo(t *testing.T) {

	svr := httptest.NewServer(http.HandlerFunc(agent))
	defer svr.Close()

	svrUrl := svr.URL
	c := DefaultClient
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

// ======= test rpc client with context ======

func TestWithXlog(t *testing.T) {

	ast := assert.New(t)
	xl := xlog.NewDummy()
	reqid := xl.ReqId()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gettedReqid := r.Header.Get("X-Reqid")
		ast.Equal(gettedReqid, reqid)
		return
	}))
	defer ts.Close()

	ctx := context.Background()
	ctx = xlog.NewContext(ctx, xl)

	resp, err := DefaultClient.DoRequest(ctx, "GET", ts.URL)
	ast.Nil(err)
	defer resp.Body.Close()
	ast.Equal(resp.StatusCode, 200)
}

func TestWithDone(t *testing.T) {

	ast := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second / 2)
	}))
	defer ts.Close()

	// test cancel
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second / 10)
		cancel()
	}()
	_, err := DefaultClient.DoRequest(ctx, "GET", ts.URL)
	ast.Equal(context.Canceled, err)

	// test timeout
	ctx, _ = context.WithTimeout(context.Background(), time.Second/10)
	_, err = DefaultClient.DoRequest(ctx, "GET", ts.URL)
	ast.Equal(context.DeadlineExceeded, err)
}

func TestCancelBeforeDo(t *testing.T) {

	ast := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("ts should not be visited")
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := DefaultClient.DoRequest(ctx, "GET", ts.URL)
	ast.Equal(context.Canceled, err)
}

func TestResponseError(t *testing.T) {

	fmtStr := "{\"error\":\"test error info\"}"
	http.HandleFunc("/ct1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf(fmtStr)))
	}))
	http.HandleFunc("/ct2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf(fmtStr)))
	}))
	http.HandleFunc("/ct3", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", " application/json ; charset=utf-8")
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf(fmtStr)))
	}))
	ts := httptest.NewServer(nil)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/ct1")
	assert.Equal(t, "test error info", ResponseError(resp).Error())
	resp, _ = http.Get(ts.URL + "/ct2")
	assert.Equal(t, "test error info", ResponseError(resp).Error())
	resp, _ = http.Get(ts.URL + "/ct3")
	assert.Equal(t, "test error info", ResponseError(resp).Error())
}
