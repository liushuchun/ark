package lb

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func Shouldretry(code int, err error) bool {
	if code == 570 {
		return true
	}
	return ShouldRetry(code, err)
}

type TestServerCfg struct {
	ExpectedBody []byte
	StatusCode   int
	ReturnBody   []byte
}

func startTestServers(t *testing.T, cfgs []*TestServerCfg) (cli *Client, servers []*httptest.Server, closer func()) {
	hosts := make([]string, 0)
	for _, c := range cfgs {
		cfg := c
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.ExpectedBody != nil {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("ReadAll r.Body failed: %v", err)
				}
				t.Logf("body: %s, expected: %s", string(body), string(cfg.ExpectedBody))
				if string(body) != string(cfg.ExpectedBody) {
					t.Fatal("io reader body cannot read again")
				}
			}
			w.WriteHeader(cfg.StatusCode)
			w.Write(cfg.ReturnBody)
		}))
		servers = append(servers, ts)
		hosts = append(hosts, ts.URL)
	}
	closer = func() {
		for _, s := range servers {
			s.Close()
		}
	}
	cli, err := New(hosts, &Config{
		TryTimes:          uint32(100),
		FailRetryInterval: -1,
		ShouldRetry:       Shouldretry,
	})
	if err != nil {
		t.Fatal("New: %v", err)
	}
	return
}

func start570server(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	cfgs := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("testtest1")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("testtest2")},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return
}

func startserver(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	expectedBody := []byte("testtesttest")
	cfgs := []*TestServerCfg{
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: []byte("testtest1")},
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: []byte("testtest2")},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return
}

func TestDo(t *testing.T) {
	xl := xlog.NewDummy()
	ast := assert.New(t)

	cli, ts, ts2 := start570server(t)
	req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ := cli.Do(xl, req)
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	closeTs(ts2)

	cli, ts, ts2 = startserver(t)
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest1", string(body))

	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}
	closeTs(ts)
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts2)
}

func TestPostWith(t *testing.T) {
	xl := xlog.NewDummy()
	ast := assert.New(t)

	cli, ts, ts2 := start570server(t)
	res, _ := cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ := ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	res, _ = cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("test post with return wrong value")
	}
	res, _ = cli.PostWith(xl, "/", "text/html", nil, 0)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	closeTs(ts2)

	cli, ts, ts2 = startserver(t)
	res, _ = cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest1" {
		t.Fatal("test post with return wrong value")
	}

	res, _ = cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	res, _ = cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))

	res, _ = cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts2)
}

func TestPostWith64(t *testing.T) {
	xl := xlog.NewDummy()
	ast := assert.New(t)

	cli, ts, ts2 := start570server(t)
	res, _ := cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ := ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	res, _ = cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	res, _ = cli.PostWith64(xl, "/", "text/html", nil, 0)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	closeTs(ts2)

	cli, ts, ts2 = startserver(t)
	res, _ = cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest1", string(body))

	res, _ = cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	res, _ = cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))

	res, _ = cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts2)
}

func TestPostWithHostRet(t *testing.T) {
	xl := xlog.NewDummy()
	ast := assert.New(t)

	cli, ts, ts2 := start570server(t)
	_, res, _ := cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ := ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", nil, 0)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	closeTs(ts2)

	cli, ts, ts2 = startserver(t)
	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest1", string(body))

	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	ast.Equal("testtest2", string(body))
	closeTs(ts)
	time.Sleep(10 * time.Millisecond)
	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))

	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest2", string(body))
	closeTs(ts2)
}

func TestPostEx(t *testing.T) {
	xl := xlog.NewDummy()
	cfgs := []*TestServerCfg{
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("testtest1")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("testtest2")},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 := servers[0], servers[1]

	res, _ := cli.PostEx(xl, "/")
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "testtest1" {
		t.Fatal("test postEx return wrong value")
	}

	res, _ = cli.PostEx(xl, "/")
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("test postEx return wrong value")
	}
	closeTs(ts)
	res, _ = cli.PostEx(xl, "/")
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("test postEx return wrong value")
	}

	res, _ = cli.PostEx(xl, "/")
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("test postEx return wrong value")
	}
	closeTs(ts2)
}

type retstruct struct {
	Val string `json:Val`
}

func start570serverjson(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	body1, _ := json.Marshal(retstruct{"testtest1"})
	body2, _ := json.Marshal(retstruct{"testtest2"})
	cfgs := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: body1},
		&TestServerCfg{StatusCode: 200, ReturnBody: body2},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return cli, ts, ts2
}

func startserverjson(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	expectedBody := []byte("testtesttest")
	body1, _ := json.Marshal(retstruct{"testtest1"})
	body2, _ := json.Marshal(retstruct{"testtest2"})
	cfgs := []*TestServerCfg{
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body1},
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body2},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return cli, ts, ts2
}

func TestCallWith(t *testing.T) {
	xl := xlog.NewDummy()

	cli, ts, ts2 := start570serverjson(t)
	var ret retstruct
	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest2" {
		t.Fatal("test post with return wrong value")
	}
	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}
	cli.CallWith(xl, &ret, "/", "text/html", nil, 0)
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}
	closeTs(ts)
	closeTs(ts2)
	cli, ts, ts2 = startserverjson(t)
	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest1" {
		t.Fatal("test call with return wrong value")
	}

	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}
	closeTs(ts)
	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}

	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}
	closeTs(ts2)
}

func TestCallWith64(t *testing.T) {
	xl := xlog.NewDummy()

	cli, ts, ts2 := start570serverjson(t)
	var ret retstruct
	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest2" {
		t.Fatal("test call with 64 return wrong value")
	}
	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest2" {
		t.Fatal("test call with 64 return wrong value")
	}
	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest2" {
		t.Fatal("test call with 64 return wrong value")
	}
	closeTs(ts)
	closeTs(ts2)
	cli, ts, ts2 = startserverjson(t)
	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest1" {
		t.Fatal("test call with return wrong value")
	}

	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}
	closeTs(ts)
	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}

	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}
	closeTs(ts2)
}

func start570serverjson2(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	expectedBody := []byte("{\"Val\":\"testtesttest\"}")
	body1, _ := json.Marshal(retstruct{"testtest1"})
	body2, _ := json.Marshal(retstruct{"testtest2"})
	cfgs := []*TestServerCfg{
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 570, ReturnBody: body1},
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body2},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return cli, ts, ts2
}

func startserverjson2(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	expectedBody := []byte("{\"Val\":\"testtesttest\"}")
	body1, _ := json.Marshal(retstruct{"testtest1"})
	body2, _ := json.Marshal(retstruct{"testtest2"})
	cfgs := []*TestServerCfg{
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body1},
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body2},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return cli, ts, ts2
}

func TestCallWithJson(t *testing.T) {
	xl := xlog.NewDummy()

	cli, ts, ts2 := start570serverjson2(t)
	var ret retstruct
	para := retstruct{"testtesttest"}
	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test post with json return wrong value")
	}
	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with json return wrong value")
	}
	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with json return wrong value")
	}
	ret = retstruct{}
	resp, err := cli.PostWithJson(xl, "/", para)
	if err != nil {
		t.Fatal(err)
	}
	if err := rpc.CallRet(xl, &ret, resp); err != nil && ret.Val != "testtest2" {
		t.Fatal(err, ret)
	}
	closeTs(ts)
	closeTs(ts2)
	cli, ts, ts2 = startserverjson2(t)
	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest1" {
		t.Fatal("test call with Json return wrong value")
	}

	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with Json return wrong value")
	}
	//	closeTs(ts)
	closeTs(ts)

	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with Json return wrong value")
	}

	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with Json return wrong value")
	}
	closeTs(ts2)
}

func closeTs(ts *httptest.Server) {
	ts.Close()
	time.Sleep(time.Millisecond)
}

func start570serverform(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	body1, _ := json.Marshal(retstruct{"testtest1"})
	body2, _ := json.Marshal(retstruct{"testtest2"})
	cfgs := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: body1},
		&TestServerCfg{StatusCode: 200, ReturnBody: body2},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return cli, ts, ts2
}

func startserverform(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	expectedBody := []byte("Val=testtesttest")
	body1, _ := json.Marshal(retstruct{"testtest1"})
	body2, _ := json.Marshal(retstruct{"testtest2"})
	cfgs := []*TestServerCfg{
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body1},
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: body2},
	}
	cli, servers, _ := startTestServers(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return cli, ts, ts2
}

func TestCallWithForm(t *testing.T) {
	xl := xlog.NewDummy()

	cli, ts, ts2 := start570serverform(t)
	var ret retstruct
	para := map[string][]string{"Val": []string{"testtesttest"}}
	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test post with form return wrong value")
	}
	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with form return wrong value")
	}
	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with form return wrong value")
	}
	ret = retstruct{}
	resp, err := cli.PostWithForm(xl, "/", para)
	if err != nil {
		t.Fatal(err)
	}
	if err := rpc.CallRet(xl, &ret, resp); err != nil && ret.Val != "testtest2" {
		t.Fatal(err, ret)
	}
	closeTs(ts)
	closeTs(ts2)

	cli, ts, ts2 = startserverform(t)
	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest1" {
		t.Fatal("test call with form return wrong value")
	}

	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with form return wrong value")
	}
	closeTs(ts)

	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with form return wrong value")
	}

	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest2" {
		t.Fatal("test call with form return wrong value")
	}
	closeTs(ts2)
}

func TestHost(t *testing.T) {

	xl := xlog.NewDummy()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if r.Host != "hosthost" {
			t.Fatal("wrong host")
		}
		w.Write([]byte("testtest1"))
	}))
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("testtest2"))
		if r.Host != "hosthost" {
			t.Fatal("wrong host")
		}
	}))
	hosts := []string{ts.URL, ts2.URL}
	cli, _ := New(hosts, &Config{
		TryTimes:          uint32(100),
		FailRetryInterval: -1,
		ShouldRetry:       Shouldretry,
	})

	req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
	req.Host = "hosthost"
	res, _ := cli.Do(xl, req)
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "testtest1" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	req.Host = "hosthost"
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}
}

func TestFailover(t *testing.T) {
	xl := xlog.NewDummy()
	var ts1, ts2, ts3, ts4 *httptest.Server
	var hosts, failoverHosts []string
	var cli *Client
	var req *Request
	var res *http.Response
	var body []byte

	ts1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(570)
		w.Write([]byte("testtest1"))
	}))
	ts2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("testtest2"))
	}))
	ts3 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("testtest3"))
	}))
	hosts = []string{ts1.URL, ts2.URL}
	failoverHosts = []string{ts3.URL}
	cli, _ = NewWithFailover(hosts, failoverHosts, &Config{
		TryTimes:          uint32(100),
		FailRetryInterval: -1,
		ShouldRetry:       Shouldretry,
	})

	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}

	ts1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(570)
		w.Write([]byte("testtest1"))
	}))
	ts2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(570)
		w.Write([]byte("testtest2"))
	}))
	ts3 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("testtest3"))
	}))
	ts4 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("testtest4"))
	}))
	hosts = []string{ts1.URL, ts2.URL}
	failoverHosts = []string{ts3.URL, ts4.URL}
	cli, _ = NewWithFailover(hosts, failoverHosts, &Config{
		TryTimes:          uint32(100),
		FailRetryInterval: -1,
		ShouldRetry:       Shouldretry,
	})
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest3" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest4" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest3" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest4" {
		t.Fatal("testdo return wrong value")
	}
}

func TestRespBody(t *testing.T) {
	xl := xlog.NewWith("TestRespBody")

	cfgsA := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("A0")},
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("A1")},
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("A2")},
	}
	_, serversA, _ := startTestServers(t, cfgsA)

	var hostsA []string
	for _, server := range serversA {
		hostsA = append(hostsA, server.URL)
	}

	// 达到重试次数，最后返回的 resp body 不应该被关闭
	cliA, err := New(hostsA, &Config{TryTimes: uint32(2), ShouldRetry: Shouldretry})
	assert.NoError(t, err)
	resp, err := cliA.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	assert.NoError(t, err)
	assert.Equal(t, 570, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 'A', body[0])
	resp.Body.Close()
}
