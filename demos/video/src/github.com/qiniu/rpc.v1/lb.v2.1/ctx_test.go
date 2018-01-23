// +build go1.5

package lb

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fmt"

	"code.google.com/p/go.net/context"
)

var wait = make(chan struct{}, 1)

func startTestServersCtx(t *testing.T, cfgs []*TestServerCfg) (cli *Client, servers []*httptest.Server, closer func()) {
	hosts := make([]string, 0)
	for _, c := range cfgs {
		cfg := c
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-wait
			w.WriteHeader(cfg.StatusCode)
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

func start570serverCtx(t *testing.T) (cli *Client, ts *httptest.Server, ts2 *httptest.Server) {
	cfgs := []*TestServerCfg{
		&TestServerCfg{StatusCode: 570, ReturnBody: []byte("testtest1")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("testtest2")},
	}
	cli, servers, _ := startTestServersCtx(t, cfgs)
	ts, ts2 = servers[0], servers[1]
	return
}

func TestCtx(t *testing.T) {
	wait = make(chan struct{}, 1)
	c := context.Background()
	ctx, cancel := context.WithCancel(c)
	ast := assert.New(t)

	cli, ts, ts2 := start570serverCtx(t)
	{
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()
		req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
		req = req.WithContext(ctx)
		res, err := cli.DoWithCtx(req)
		close(wait)
		ast.Error(err)
		ast.Equal(ctx.Err(), context.Canceled)
		if runtime.Version() >= "go1.6" {
			ast.True(strings.Contains(err.Error(), "net/http: request canceled"))
		}
		if res != nil {
			res.Body.Close()
		}
	}
	closeTs(ts)
	closeTs(ts2)
}

func TestCtxRequestCancel(t *testing.T) {
	ctx := context.Background()
	ast := assert.New(t)
	wait = make(chan struct{}, 1)

	cli, ts, ts2 := start570serverCtx(t)
	{
		req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
		c := make(chan struct{})
		req.Cancel = c
		go func() {
			time.Sleep(100 * time.Millisecond)
			close(c)
		}()
		req = req.WithContext(ctx)
		res, err := cli.DoWithCtx(req)
		close(wait)
		ast.Error(err)
		ast.NoError(ctx.Err())
		ast.True(strings.Contains(err.Error(), "net/http: request canceled"))
		if res != nil {
			res.Body.Close()
		}
	}
	closeTs(ts)
	closeTs(ts2)
}

func TestCancelRequestWithChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in -short mode")
	}
	unblockc := make(chan bool)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
		w.(http.Flusher).Flush() // send headers and some body
		<-unblockc
	}))
	defer ts.Close()
	defer close(unblockc)

	tr := NewTransport(nil).(*Transport)
	defer tr.tr.CloseIdleConnections()
	c := New(&Config{Hosts: []string{ts.URL}}, tr)

	req, _ := NewRequest("GET", ts.URL, nil)
	ch := make(chan struct{})
	req.Cancel = ch

	res, err := c.DoWithCtx(req)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		close(ch)
	}()
	t0 := time.Now()
	body, err := ioutil.ReadAll(res.Body)
	d := time.Since(t0)

	if err.Error() != "net/http: request canceled" {
		t.Errorf("Body.Read error = %v; want errRequestCanceled", err)
	}
	if string(body) != "Hello" {
		t.Errorf("Body = %q; want Hello", body)
	}
	if d < 500*time.Millisecond || d > 1500*time.Millisecond {
		t.Errorf("expected ~1 second delay; got %v", d)
	}
}
