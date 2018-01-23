package lb

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"

	"qbox.us/rateio"
	"sync/atomic"
)

func init() {
	log.SetOutputLevel(0)
}

func shouldRetry(code int, err error) bool {
	if code == 570 {
		return true
	}
	return ShouldRetry(code, err)
}

func shouldReproxy(code int, err error) bool {
	if code == 575 {
		return true
	}
	return ShouldReproxy(code, err)
}

func shouldFailover(code int, err error) bool {
	return shouldRetry(code, err) || shouldReproxy(code, err)
}

type TestServerCfg struct {
	ExpectedBody []byte
	StatusCode   int
	ReturnBody   []byte
	Rate         int

	RespHeaderTime int
}

func startTestServers(t *testing.T, cfgs []*TestServerCfg) (cli *Client, servers []*httptest.Server, closer func()) {
	hosts := make([]string, 0)
	for _, c := range cfgs {
		cfg := c
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trytime++
			if c.RespHeaderTime > 0 {
				time.Sleep(time.Second * time.Duration(c.RespHeaderTime))
			}
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
			if cfg.Rate != 0 {
				io.Copy(w, rateio.NewRateReader(bytes.NewBuffer(cfg.ReturnBody), cfg.Rate))
			} else {
				w.Write(cfg.ReturnBody)
			}
		}))
		servers = append(servers, ts)
		hosts = append(hosts, ts.URL)
	}
	closer = func() {
		for _, s := range servers {
			s.Close()
		}
	}
	cli = New(&Config{
		Hosts:       hosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,
	}, nil)
	return
}

func startTestServers2(t *testing.T, cfgs []*TestServerCfg) (cli *Client, servers []*httptest.Server, closer func()) {
	hosts := make([]string, 0)
	for _, c := range cfgs {
		cfg := c
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll r.Body failed: %v", err)
			}
			t.Logf("body: %s, expected: %s", string(body), string(cfg.ExpectedBody))
			if cfg.ExpectedBody != nil {
				if string(body) != string(cfg.ExpectedBody) {
					t.Fatal("io reader body cannot read again")
				}
			}
			w.WriteHeader(cfg.StatusCode)
			if cfg.Rate != 0 {
				io.Copy(w, rateio.NewRateReader(bytes.NewBuffer(cfg.ReturnBody), cfg.Rate/2))
			} else {
				w.Write(cfg.ReturnBody)
			}
		}))
		servers = append(servers, ts)
		hosts = append(hosts, ts.URL)
	}
	closer = func() {
		for _, s := range servers {
			s.Close()
		}
	}
	cli = New(&Config{
		Hosts:       hosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,

		FailRetryIntervalS: 5,
		MaxFails:           2,
		MaxFailsPeriodS:    1,
		SpeedLimit: SpeedLimit{
			CalcSpeedSizeThresholdB: 10,
			BanHostBelowBps:         1024*1024 + 1,
		},
	}, nil)
	return
}

var proxycount uint64

type proxyServer struct {
	id int
}

func (s *proxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var hopHeaders = []string{"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
		"Te", "Trailers", "Transfer-Encoding", "Upgrade", "Proxy-Connection"}

	xl := xlog.NewWithReq(r)
	xl.Info(r)
	atomic.AddUint64(&proxycount, 1)
	if !r.URL.IsAbs() {
		xl.Panic("not proxy request")
	}
	for _, k := range hopHeaders {
		r.Header.Del(k)
	}
	tr := http.DefaultTransport
	resp, err := tr.RoundTrip(r)
	if err != nil {
		httputil.Error(w, httputil.NewError(502, err.Error()))
		return
	}
	defer resp.Body.Close()
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("X-ProxyIndex", strconv.FormatInt(int64(s.id), 10))
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		xl.Error(err)
	}
}

func startProxys(n int) (servers []*httptest.Server, closer func()) {
	for i := 0; i < n; i++ {
		ts := httptest.NewServer(&proxyServer{i})
		servers = append(servers, ts)
	}
	closer = func() {
		for _, s := range servers {
			s.Close()
		}
	}
	return
}

var returnLongBody1 = make([]byte, 64*1024)
var returnLongBody2 = make([]byte, 64*1024)

func startTLEserver(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	returnLongBody1[0] = byte(1)
	returnLongBody2[0] = byte(2)
	expectedBody := []byte("testtesttest")
	cfgs := []*TestServerCfg{
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: returnLongBody1},
		&TestServerCfg{ExpectedBody: expectedBody, StatusCode: 200, ReturnBody: returnLongBody2, Rate: 1024 * 1024},
	}
	cli, servers, _ := startTestServers2(t, cfgs)
	ts, ts2 = servers[0], servers[1]
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
	ast.Equal("testtest2", string(body))

	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest1" {
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

func TestDo2(t *testing.T) {
	xl := xlog.NewDummy()
	ast := assert.New(t)

	cli, ts, ts2 := startTLEserver(t)
	req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ := cli.Do(xl, req)
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != string(returnLongBody2) {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal(string(returnLongBody1), string(body))
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal(returnLongBody2, body)
	time.Sleep(time.Second)
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal(string(returnLongBody1), string(body))
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal(returnLongBody1, body)
	time.Sleep(time.Second)
	closeTs(ts)
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
	if string(body) != "testtest2" {
		t.Fatal("test post with return wrong value")
	}

	res, _ = cli.PostWith(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest1", string(body))
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
	ast.Equal("testtest2", string(body))

	res, _ = cli.PostWith64(xl, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	body, _ = ioutil.ReadAll(res.Body)
	ast.Equal("testtest1", string(body))
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
	ast.Equal("testtest2", string(body))

	_, res, _ = cli.PostWithHostRet(xl, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	body, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	ast.Equal("testtest1", string(body))
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
	if string(body) != "testtest2" {
		t.Fatal("test postEx return wrong value")
	}

	res, _ = cli.PostEx(xl, "/")
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest1" {
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
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}

	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest1" {
		t.Fatal("test call with return wrong value")
	}
	closeTs(ts)
	cli.CallWith(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), len("testtesttest"))
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value", ret.Val)
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
	if ret.Val != "testtest2" {
		t.Fatal("test call with return wrong value")
	}

	cli.CallWith64(xl, &ret, "/", "text/html", strings.NewReader("testtesttest"), int64(len("testtesttest")))
	if ret.Val != "testtest1" {
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
	if ret.Val != "testtest2" {
		t.Fatal("test call with Json return wrong value")
	}

	cli.CallWithJson(xl, &ret, "/", para)
	if ret.Val != "testtest1" {
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
	if ret.Val != "testtest2" {
		t.Fatal("test call with form return wrong value")
	}

	cli.CallWithForm(xl, &ret, "/", para)
	if ret.Val != "testtest1" {
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
	cli := New(&Config{
		Hosts:       hosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,
	}, nil)

	req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
	req.Host = "hosthost"
	res, _ := cli.Do(xl, req)
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "testtest2" {
		t.Fatal("testdo return wrong value")
	}
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	req.Host = "hosthost"
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest1" {
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
	cli = NewWithFailover(&Config{
		Hosts:       hosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,
	}, &Config{
		Hosts:       failoverHosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,
	}, nil, nil, shouldFailover)

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
	cli = NewWithFailover(&Config{
		Hosts:       hosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,
	}, &Config{
		Hosts:       failoverHosts,
		TryTimes:    uint32(10),
		ShouldRetry: shouldRetry,
	}, nil, nil, shouldFailover)
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
	req, _ = NewRequest("POST", "/", strings.NewReader("testtesttest"))
	res, _ = cli.Do(xl, req)
	body, _ = ioutil.ReadAll(res.Body)
	if string(body) != "testtest3" {
		t.Fatal("testdo return wrong value")
	}
}

func TestProxy(t *testing.T) {
	xl := xlog.NewWith("TestProxy")

	cfgsA := []*TestServerCfg{
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("A0")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("A1")},
	}
	cfgsB := []*TestServerCfg{
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("B0")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("B1")},
	}
	_, serversA, _ := startTestServers(t, cfgsA)
	_, serversB, _ := startTestServers(t, cfgsB)

	proxys, _ := startProxys(2)

	var hostsA []string // 不通过代理访问
	for _, server := range serversA {
		hostsA = append(hostsA, server.URL)
	}
	var hostsB []string // 通过代理访问
	for _, server := range serversB {
		hostsB = append(hostsB, server.URL)
	}
	var proxyHostsB []string
	for _, server := range proxys {
		proxyHostsB = append(proxyHostsB, server.URL)
	}

	trB := NewTransport(&TransportConfig{
		Proxys:        proxyHostsB,
		TryTimes:      uint32(10), // > 2x2
		ShouldReproxy: shouldReproxy,
	})

	cli := NewWithFailover(&Config{
		Hosts:       hostsA,
		TryTimes:    uint32(10), // > 2x2
		ShouldRetry: shouldRetry,
	}, &Config{
		Hosts:       hostsB,
		TryTimes:    uint32(10), // > 2x2
		ShouldRetry: shouldRetry,
	}, nil, trB, shouldFailover)

	//成功请求
	xl.Info("Case OK")
	for i := 0; i < 2*2; i++ {
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		ebody := "A" + strconv.FormatInt(int64(i+1)%2, 10) // A1, A0, ...
		assert.Equal(t, ebody, body)                       // 轮询
		assert.Empty(t, resp.Header.Get("X-ProxyIndex"))   // 不通过 proxy
		resp.Body.Close()
	}

	// 停止 A1
	xl.Info("Case STOP A1")
	serversA[1].Close()
	for i := 0; i < 2*2; i++ {
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "A0", string(body), "%v %s", i, body) // 轮询
		assert.Empty(t, resp.Header.Get("X-ProxyIndex"))      // 不通过 proxy
		resp.Body.Close()
	}

	// 所有 A 组都停止，会 failover 到 B 组有代理
	xl.Info("Case STOP A")
	for _, server := range serversA {
		server.Close()
	}
	for i := 0; i < 2*2; i++ {
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		ebody := "B" + strconv.FormatInt(int64(i+1)%2, 10)                  //
		assert.Equal(t, ebody, string(body), "%v %s != %s", i, ebody, body) // 轮询
		assert.NotEmpty(t, resp.Header.Get("X-ProxyIndex"))                 // 通过 proxy
		resp.Body.Close()
	}

	// 停止 B1，会有 502 自动重试
	xl.Info("Case STOP B1")
	serversB[1].Close()
	for i := 0; i < 2*2; i++ {
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "B0", string(body), "%v %s", i, body) // 轮询
		assert.NotEmpty(t, resp.Header.Get("X-ProxyIndex"))   // 通过 proxy
		resp.Body.Close()
	}

	// 停止代理 1，可以正常访问
	xl.Info("Case STOP P1")
	proxys[1].Close()
	for i := 0; i < 2*2; i++ {
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "B0", string(body), "%v %s", i, body) // 轮询
		assert.Equal(t, "0", resp.Header.Get("X-ProxyIndex"))
		resp.Body.Close()
	}

	// 所有代理都挂了，错误
	xl.Info("Case STOP P")
	for _, p := range proxys {
		p.Close()
	}
	_, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	xl.Error(err, httputil.DetectCode(err))
	assert.Error(t, err)
}

var trytime uint64

func TestProxy2(t *testing.T) {
	xl := xlog.NewWith("TestProxy")
	trytime = 0

	cfgsA := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("A1")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("A0")},
	}
	cfgsB := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("B1")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("B0")},
	}
	_, serversA, _ := startTestServers(t, cfgsA)
	_, serversB, _ := startTestServers(t, cfgsB)

	proxys, _ := startProxys(2)

	var hostsA []string // 不通过代理访问
	for _, server := range serversA {
		hostsA = append(hostsA, server.URL)
	}
	var hostsB []string // 通过代理访问
	for _, server := range serversB {
		hostsB = append(hostsB, server.URL)
	}
	var proxyHostsB []string
	for _, server := range proxys {
		proxyHostsB = append(proxyHostsB, server.URL)
	}

	trB := NewTransport(&TransportConfig{
		Proxys:        proxyHostsB,
		TryTimes:      uint32(10), // > 2x2
		ShouldReproxy: shouldReproxy,
	})

	cli := NewWithFailover(&Config{
		Hosts:           hostsA,
		TryTimes:        uint32(10), // > 2x2
		ShouldRetry:     shouldRetry,
		MaxFails:        2,
		MaxFailsPeriodS: 1,

		FailRetryIntervalS: 10,
	}, &Config{
		Hosts:           hostsB,
		TryTimes:        uint32(10), // > 2x2
		ShouldRetry:     shouldRetry,
		MaxFails:        2,
		MaxFailsPeriodS: 1,

		FailRetryIntervalS: 10,
	}, nil, trB, shouldFailover)

	//成功请求
	xl.Info("Case OK")
	time.Sleep(1 * time.Second)
	for i := 0; i < 2*2; i++ {
		trytime = 0
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		ebody := "A0"
		assert.Equal(t, ebody, body)                     // 轮询
		assert.Empty(t, resp.Header.Get("X-ProxyIndex")) // 不通过 proxy
		resp.Body.Close()
		assert.Equal(t, trytime, i%2+1, "%v %d", i, trytime)
	}
	for i := 0; i < 2; i++ {
		trytime = 0
		cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.Equal(t, trytime, 1, "%v %d", 1, trytime)
	}

	// 停止 A1
	xl.Info("Case STOP A1")
	serversA[0].Close()
	for i := 0; i < 2*2; i++ {
		trytime = 0
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "A0", string(body), "%v %s", i, body) // 轮询
		assert.Empty(t, resp.Header.Get("X-ProxyIndex"))      // 不通过 proxy
		resp.Body.Close()
		assert.Equal(t, trytime, 1, "%v %d", i, trytime)
	}

	// 所有 A 组都停止，会 failover 到 B 组有代理
	xl.Info("Case STOP A")
	for _, server := range serversA {
		server.Close()
	}
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 2*2; i++ {
		trytime = 0
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		ebody := "B0"                                                       //
		assert.Equal(t, ebody, string(body), "%v %s != %s", i, ebody, body) // 轮询
		assert.NotEmpty(t, resp.Header.Get("X-ProxyIndex"))                 // 通过 proxy
		resp.Body.Close()
		assert.Equal(t, trytime, i%2+1, "%d %d", i, trytime)
	}
	return

	// 停止 B1，会有 502 自动重试
	xl.Info("Case STOP B1")
	serversB[0].Close()
	for i := 0; i < 2*2; i++ {
		trytime = 0
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "B0", string(body), "%v %s", i, body) // 轮询
		assert.NotEmpty(t, resp.Header.Get("X-ProxyIndex"))   // 通过 proxy
		resp.Body.Close()
		assert.Equal(t, trytime, 1, "%v %d", i, trytime)
	}

	// 停止代理 1，可以正常访问
	xl.Info("Case STOP P1")
	proxys[1].Close()
	for i := 0; i < 2*2; i++ {
		trytime = 0
		resp, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "B0", string(body), "%v %s", i, body) // 轮询
		assert.Equal(t, "0", resp.Header.Get("X-ProxyIndex"))
		resp.Body.Close()
		assert.Equal(t, trytime, 1, "%v %d", i, trytime)
	}

	// 所有代理都挂了，错误
	xl.Info("Case STOP P")
	for _, p := range proxys {
		p.Close()
	}
	_, err := cli.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	xl.Error(err, httputil.DetectCode(err))
	assert.Error(t, err)
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

	// 1. 达到重试次数，最后返回的 resp body 不应该被关闭
	xl.Info("CASE: reach try times and failed")
	cliA := New(&Config{Hosts: hostsA, TryTimes: uint32(2), ShouldRetry: shouldRetry}, nil)
	resp, err := cliA.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	assert.NoError(t, err)
	assert.Equal(t, 570, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 'A', body[0])
	resp.Body.Close()

	// 2. 未达到重试次数，但是 host 或者 proxy 全部被屏蔽了，resp body 也不应该被关闭
	xl.Info("CASE: service unavailable")
	cliB := New(&Config{Hosts: hostsA, TryTimes: uint32(10), ShouldRetry: shouldRetry}, nil)
	resp, err = cliB.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	assert.NoError(t, err)
	assert.Equal(t, 570, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 'A', body[0])
	resp.Body.Close()

	// 测试代理

	cfgsB := []*TestServerCfg{
		&TestServerCfg{StatusCode: 575, ReturnBody: []byte("A0")},
		&TestServerCfg{StatusCode: 575, ReturnBody: []byte("A1")},
		&TestServerCfg{StatusCode: 575, ReturnBody: []byte("A2")},
	}
	_, serversB, _ := startTestServers(t, cfgsB)
	var hostsB []string
	for _, server := range serversB {
		hostsB = append(hostsB, server.URL)
	}

	proxys, _ := startProxys(3)
	var proxyHosts []string
	for _, server := range proxys {
		proxyHosts = append(proxyHosts, server.URL)
	}

	xl.Info("CASE: proxy reach try times and failed")
	trB := NewTransport(&TransportConfig{Proxys: proxyHosts, TryTimes: uint32(2), ShouldReproxy: shouldReproxy})
	cliAp := New(&Config{Hosts: hostsB, TryTimes: uint32(2), ShouldRetry: shouldRetry}, trB)
	resp, err = cliAp.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	assert.NoError(t, err)
	assert.Equal(t, 575, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 'A', body[0])
	resp.Body.Close()

	// 2. 未达到重试次数，但是 host 或者 proxy 全部被屏蔽了，resp body 也不应该被关闭
	xl.Info("CASE: proxy service unavailable")
	trB1 := NewTransport(&TransportConfig{Proxys: proxyHosts, TryTimes: uint32(2), ShouldReproxy: shouldReproxy})
	cliBp := New(&Config{Hosts: hostsB, TryTimes: uint32(10), ShouldRetry: shouldRetry}, trB1)
	resp, err = cliBp.PostWith(xl, "/", "application/octet-stream", strings.NewReader(""), 0)
	assert.NoError(t, err)
	assert.Equal(t, 575, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 'A', body[0])
	resp.Body.Close()

}

var alwaysTrue = func(code int, err error) bool { return true }

func TestDnsResolveLB(t *testing.T) {
	countA, countB := 0, 0
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countA++
		w.WriteHeader(200)
	}))
	defer a.Close()
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countB++
		w.WriteHeader(200)
	}))
	b.Close()
	time.Sleep(1e9)

	lookupHost := func(host string) (addrs []string, err error) {
		l := len("http://")
		return []string{a.URL[l:]}, nil
	}
	LookupHost := func(host string) (addrs []string, err error) {
		return lookupHost(host)
	}
	cli1 := New(&Config{
		Hosts:         []string{"http://foo"},
		DnsResolve:    true,
		DnsCacheTimeS: 1,
		LookupHost:    LookupHost,
	}, nil)
	cli2 := New(&Config{
		Hosts: []string{b.URL}, // closed
	}, NewTransport(&TransportConfig{
		Proxys:        []string{"http://foo"},
		DnsResolve:    true,
		DnsCacheTimeS: 1,
		LookupHost:    LookupHost,
	}))
	for i := 0; i < 10; i++ {
		err := cli1.Call(nil, nil, "/bar")
		assert.NoError(t, err)
		err = cli2.Call(nil, nil, "/bar")
		assert.NoError(t, err)
	}
	assert.Equal(t, 20, countA)

	lookupHost = func(host string) (addrs []string, err error) {
		l := len("http://")
		return []string{a.URL[l:], b.URL[l:]}, nil
	}
	time.Sleep(1.1e9)
	countA = 0
	for i := 0; i < 10; i++ {
		err := cli1.Call(nil, nil, "/bar")
		assert.NoError(t, err)
		err = cli2.Call(nil, nil, "/bar")
		assert.NoError(t, err)
	}
	assert.Equal(t, 20, countA)
}

func TestDnsResolveSingle(t *testing.T) {
	countA, countB := 0, 0
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countA++
		assert.False(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, "domainA")
		w.WriteHeader(801)
	}))
	defer a.Close()
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countB++
		assert.True(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, "domainA")
		w.WriteHeader(802)
	}))
	defer b.Close()

	lookupHost := func(host string) (addrs []string, err error) {
		l := len("http://")
		if host == "domainA" {
			return []string{a.URL[l:]}, nil
		}
		if host == "domainB" {
			return []string{b.URL[l:]}, nil
		}
		panic("cannot reach here")
	}
	LookupHost := func(host string) (addrs []string, err error) {
		return lookupHost(host)
	}
	{
		countA, countB = 0, 0
		cli := NewWithFailover(
			&Config{
				Hosts:      []string{"http://domainA"},
				DnsResolve: true,
				LookupHost: LookupHost,
			},
			&Config{
				Hosts:      []string{"http://domainA"},
				DnsResolve: true,
				LookupHost: LookupHost,
			},
			nil,
			NewTransport(&TransportConfig{
				Proxys:     []string{"http://domainB"},
				DnsResolve: true,
				LookupHost: LookupHost,
			}),
			alwaysTrue,
		)
		err := cli.Call(nil, nil, "/")
		assert.Equal(t, 802, httputil.DetectCode(err))
		assert.Equal(t, 1, countA)
		assert.Equal(t, 1, countB)
	}
	{
		countA, countB = 0, 0
		cli := NewWithFailover(
			&Config{
				Hosts:      []string{"http://domainA"},
				DnsResolve: true,
				LookupHost: LookupHost,
			},
			&Config{
				Hosts:      []string{"http://domainA"},
				DnsResolve: true,
				LookupHost: LookupHost,
			},
			NewTransport(&TransportConfig{
				Proxys:     []string{"http://domainB"},
				DnsResolve: true,
				LookupHost: LookupHost,
			}),
			NewTransport(&TransportConfig{
				Proxys:     []string{"http://domainB"},
				DnsResolve: true,
				LookupHost: LookupHost,
			}),
			alwaysTrue,
		)
		err := cli.Call(nil, nil, "/")
		assert.Equal(t, 802, httputil.DetectCode(err))
		assert.Equal(t, 0, countA)
		assert.Equal(t, 2, countB)
	}

	countA2, countB2 := 0, 0
	a2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countA2++
		assert.False(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, "domainA")
		w.WriteHeader(801)
	}))
	defer a2.Close()
	b2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countB2++
		assert.True(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, "domainA")
		w.WriteHeader(802)
	}))
	defer b2.Close()
	tr := NewTransport(&TransportConfig{
		Proxys:        []string{"http://domainB"},
		DnsResolve:    true,
		DnsCacheTimeS: 1,
		LookupHost:    LookupHost,
	})
	cli := NewWithFailover(
		&Config{
			Hosts:         []string{"http://domainA"},
			DnsResolve:    true,
			DnsCacheTimeS: 1,
			LookupHost:    LookupHost,
		},
		&Config{
			Hosts:         []string{"http://domainA"},
			DnsResolve:    true,
			DnsCacheTimeS: 1,
			LookupHost:    LookupHost,
		},
		tr,
		nil,
		alwaysTrue,
	)
	{
		countA, countB = 0, 0
		countA2, countB2 = 0, 0
		err := cli.Call(nil, nil, "/")
		assert.Equal(t, 801, httputil.DetectCode(err))
		assert.Equal(t, 1, countA)
		assert.Equal(t, 1, countB)
		assert.Equal(t, 0, countA2)
		assert.Equal(t, 0, countB2)
	}
	lookupHost = func(host string) (addrs []string, err error) {
		//return []string{}, nil
		l := len("http://")
		if host == "domainA" {
			return []string{a2.URL[l:]}, nil
		}
		if host == "domainB" {
			return []string{b2.URL[l:]}, nil
		}
		panic("cannot reach here" + host)
	}

	err := cli.client.sel.resolveDns()
	assert.NoError(t, err)
	err = cli.failover.sel.resolveDns()
	assert.NoError(t, err)
	err = tr.(*Transport).sel.resolveDns()
	assert.NoError(t, err)
	time.Sleep(1.2e9)
	{
		countA, countB = 0, 0
		countA2, countB2 = 0, 0
		err := cli.Call(nil, nil, "/")
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 801, httputil.DetectCode(err))
		assert.Equal(t, 0, countA)
		assert.Equal(t, 0, countB)
		assert.Equal(t, 1, countA2)
		assert.Equal(t, 1, countB2)
	}
}

func TestLookupNotHoldHost(t *testing.T) {
	countA, countB := 0, 0
	var urlA string
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countA++
		assert.False(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, urlA[7:])
		w.WriteHeader(801)
	}))
	defer a.Close()
	urlA = a.URL
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countB++
		assert.True(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, a.URL[7:])
		w.WriteHeader(802)
	}))
	defer b.Close()

	lookupHost := func(host string) (addrs []string, err error) {
		l := len("http://")
		if host == "domainA" {
			return []string{a.URL[l:]}, nil
		}
		if host == "domainB" {
			return []string{b.URL[l:]}, nil
		}
		panic("cannot reach here")
	}
	LookupHost := func(host string) (addrs []string, err error) {
		return lookupHost(host)
	}
	{
		countA, countB = 0, 0
		cli := NewWithFailover(
			&Config{
				Hosts:                 []string{"http://domainA"},
				DnsResolve:            true,
				LookupHost:            LookupHost,
				LookupHostNotHoldHost: true,
			},
			&Config{
				Hosts:                 []string{"http://domainA"},
				DnsResolve:            true,
				LookupHost:            LookupHost,
				LookupHostNotHoldHost: true,
			},
			nil,
			NewTransport(&TransportConfig{
				Proxys:     []string{"http://domainB"},
				DnsResolve: true,
				LookupHost: LookupHost,
			}),
			alwaysTrue,
		)
		err := cli.Call(nil, nil, "/")
		assert.Equal(t, 802, httputil.DetectCode(err))
		assert.Equal(t, 1, countA)
		assert.Equal(t, 1, countB)
	}
	{
		countA, countB = 0, 0
		cli := NewWithFailover(
			&Config{
				Hosts:                 []string{"http://domainA"},
				DnsResolve:            true,
				LookupHost:            LookupHost,
				LookupHostNotHoldHost: true,
			},
			&Config{
				Hosts:                 []string{"http://domainA"},
				DnsResolve:            true,
				LookupHost:            LookupHost,
				LookupHostNotHoldHost: true,
			},
			NewTransport(&TransportConfig{
				Proxys:     []string{"http://domainB"},
				DnsResolve: true,
				LookupHost: LookupHost,
			}),
			NewTransport(&TransportConfig{
				Proxys:     []string{"http://domainB"},
				DnsResolve: true,
				LookupHost: LookupHost,
			}),
			alwaysTrue,
		)
		err := cli.Call(nil, nil, "/")
		assert.Equal(t, 802, httputil.DetectCode(err))
		assert.Equal(t, 0, countA)
		assert.Equal(t, 2, countB)
	}
	var urlA2 string
	countA2, countB2 := 0, 0
	a2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countA2++
		assert.False(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, urlA2[7:])
		w.WriteHeader(801)
	}))
	defer a2.Close()
	urlA2 = a2.URL
	b2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		countB2++
		assert.True(t, req.URL.IsAbs())
		assert.Equal(t, req.Host, a2.URL[7:])
		w.WriteHeader(802)
	}))
	defer b2.Close()
	tr := NewTransport(&TransportConfig{
		Proxys:        []string{"http://domainB"},
		DnsResolve:    true,
		DnsCacheTimeS: 1,
		LookupHost:    LookupHost,
	})
	cli := NewWithFailover(
		&Config{
			Hosts:                 []string{"http://domainA"},
			DnsResolve:            true,
			DnsCacheTimeS:         1,
			LookupHost:            LookupHost,
			LookupHostNotHoldHost: true,
		},
		&Config{
			Hosts:                 []string{"http://domainA"},
			DnsResolve:            true,
			DnsCacheTimeS:         1,
			LookupHost:            LookupHost,
			LookupHostNotHoldHost: true,
		},
		tr,
		nil,
		alwaysTrue,
	)
	{
		countA, countB = 0, 0
		countA2, countB2 = 0, 0
		err := cli.Call(nil, nil, "/")
		assert.Equal(t, 801, httputil.DetectCode(err))
		assert.Equal(t, 1, countA)
		assert.Equal(t, 1, countB)
		assert.Equal(t, 0, countA2)
		assert.Equal(t, 0, countB2)
	}
	lookupHost = func(host string) (addrs []string, err error) {
		//return []string{}, nil
		l := len("http://")
		if host == "domainA" {
			return []string{a2.URL[l:]}, nil
		}
		if host == "domainB" {
			return []string{b2.URL[l:]}, nil
		}
		panic("cannot reach here" + host)
	}

	err := cli.client.sel.resolveDns()
	assert.NoError(t, err)
	err = cli.failover.sel.resolveDns()
	assert.NoError(t, err)
	err = tr.(*Transport).sel.resolveDns()
	assert.NoError(t, err)
	time.Sleep(1.2e9)
	{
		countA, countB = 0, 0
		countA2, countB2 = 0, 0
		err := cli.Call(nil, nil, "/")
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 801, httputil.DetectCode(err))
		assert.Equal(t, 0, countA)
		assert.Equal(t, 0, countB)
		assert.Equal(t, 1, countA2)
		assert.Equal(t, 1, countB2)
	}
}
