package lb

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/stretchr/testify/assert"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

type mockSvr struct {
	Bodys map[string][][]byte
	Code  int
	Wait  *sync.WaitGroup
	Done  *sync.WaitGroup
}

func openMockSvr() (*mockSvr, *httptest.Server) {

	s := &mockSvr{
		Bodys: make(map[string][][]byte),
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if s.Done != nil {
			defer s.Done.Done()
		}

		xl := xlog.New(w, req)
		xl.Infof("mocksvr: %#v %v %v", w, req.Host, req.RemoteAddr)

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			httputil.ReplyErr(w, 400, "ReadAll:"+err.Error())
			return
		}
		if s.Wait != nil {
			s.Wait.Wait()
		}
		if s.Code > 0 {
			httputil.ReplyErr(w, s.Code, "explicit")
			return
		}
		path := req.URL.Path[1:]
		s.Bodys[path] = append(s.Bodys[path], body)
		httputil.Reply(w, 200, map[string]string{
			"testing": "ok",
		})
	}))
	return s, ts
}

func openMockSvrs(n int) ([]*mockSvr, []*httptest.Server) {

	ss := make([]*mockSvr, n)
	tss := make([]*httptest.Server, n)
	for i := 0; i < n; i++ {
		ss[i], tss[i] = openMockSvr()
	}
	return ss, tss
}

func closeServers(tss []*httptest.Server) {

	for _, ts := range tss {
		ts.Close()
	}
}

func collectHosts(tss []*httptest.Server) []string {

	hosts := make([]string, len(tss))
	for i, ts := range tss {
		hosts[i] = ts.URL
	}
	return hosts
}

func TestNormal(t *testing.T) {

	svrs, tss := openMockSvrs(3)

	cfg := &Config{
		Hosts:      collectHosts(tss),
		HostRetrys: 2,
	}
	c := New(cfg)

	ret := make(map[string]string)
	ctx := context.Background()
	data := []byte("helloworld")

	// normal
	for i := 0; i < 3; i++ {
		r := bytes.NewReader(data)
		err := c.CallWith(ctx, &ret, "/good", "", r, r.Len())
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"testing": "ok"}, ret)
		var count int
		for _, svr := range svrs {
			if datas, ok := svr.Bodys["good"]; ok {
				assert.Equal(t, 1, len(datas))
				assert.Equal(t, data, datas[0])
				count++
			}
		}
		assert.Equal(t, i+1, count)
	}

	// 1/3 400
	count := 0
	svrs[0].Code = 400
	for i := 0; i < 3; i++ {
		r := bytes.NewReader(data)
		err := c.CallWith(ctx, &ret, "/bad", "", r, r.Len())
		if e, ok := err.(rpc.RespError); ok && e.HttpCode() == 400 {
			count++
		} else {
			assert.NoError(t, err)
			assert.Equal(t, map[string]string{"testing": "ok"}, ret)
		}
	}
	assert.Equal(t, 1, count)

	// 1/3 closed
	tss[0].Close()
	for i := 0; i < 3; i++ {
		r := bytes.NewReader(data)
		err := c.CallWith(ctx, &ret, "/good", "", r, r.Len())
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"testing": "ok"}, ret)
	}

	// 1/3 closed and 1/3 570
	svrs[1].Code = 570
	for i := 0; i < 3; i++ {
		r := bytes.NewReader(data)
		err := c.CallWith(ctx, &ret, "/good", "", r, r.Len())
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"testing": "ok"}, ret)
	}

	// 2/3 closed and 1/3 570
	tss[2].Close()
	for i := 0; i < 3; i++ {
		r := bytes.NewReader(data)
		err := c.CallWith(ctx, &ret, "/5xx", "", r, r.Len())
		if e, ok := err.(rpc.RespError); ok {
			assert.Equal(t, 570, e.HttpCode())
		} else {
			// 由于 tss[N] 都被 Close 了，对应的端口可能被重新分配
			// 如果上述情况出现，有可能出现非 connection refused 的错误
			// 在 go1.6 底下确实遇到过一次 EOF 错误，排查无结果，猜测是上述原因导致
			if !strings.Contains(err.Error(), "connection refused") {
				xl := xlog.NewDummy()
				xl.Warnf("%d unexpected error %+v", i, err)
				assert.Contains(t, err.Error(), "EOF")
			}
		}
	}

	closeServers(tss)
}

func TestCancel(t *testing.T) {

	svrs, tss := openMockSvrs(3)
	cfg := &Config{
		Hosts:      collectHosts(tss),
		HostRetrys: 2,
	}
	c := New(cfg)

	data := []byte("helloworld")
	r := bytes.NewReader(data)
	ret := make(map[string]string)
	ctx, canceler := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for _, svr := range svrs {
		svr.Wait = &wg
	}
	wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		canceler()
		wg.Done()
	}()
	err := c.CallWith(ctx, &ret, "/cancel", "", r, r.Len())
	assert.Equal(t, ErrCanceled, err)

	closeServers(tss)
}

func TestCancelWith5xx(t *testing.T) {

	svrs, tss := openMockSvrs(3)
	cfg := &Config{
		Hosts:      collectHosts(tss),
		HostRetrys: 2,
	}
	c := New(cfg)

	data := []byte("helloworld")
	r := bytes.NewReader(data)
	ret := make(map[string]string)
	ctx, canceler := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	svrs[0].Code = 570
	svrs[1].Code = 570
	svrs[2].Wait = &wg
	wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		canceler()
		wg.Done()
	}()
	err := c.CallWith(ctx, &ret, "/570", "", r, r.Len())
	assert.Equal(t, ErrCanceled, err)

	closeServers(tss)
}

func TestTimeoutRetry(t *testing.T) {

	svrs, tss := openMockSvrs(3)
	cfg := &Config{
		Hosts:          collectHosts(tss),
		HostRetrys:     2,
		RetryTimeoutMs: 10,
	}
	c := New(cfg)

	data := []byte("helloworld")
	r := bytes.NewReader(data)
	ret := make(map[string]string)
	ctx := context.Background()

	var wait sync.WaitGroup
	var done sync.WaitGroup
	for _, svr := range svrs {
		svr.Wait = &wait
		svr.Done = &done
	}
	wait.Add(1)
	done.Add(3)
	go func() {
		time.Sleep(100 * time.Millisecond)
		wait.Done()
	}()
	err := c.CallWith(ctx, &ret, "/200", "", r, r.Len())
	assert.NoError(t, err)
	done.Wait()
	var count int
	for _, svr := range svrs {
		if datas, ok := svr.Bodys["200"]; ok {
			assert.Equal(t, 1, len(datas))
			assert.Equal(t, data, datas[0])
			count++
		}
	}
	assert.Equal(t, 3, count)
}

func TestTimeoutRetry5xx(t *testing.T) {

	svrs, tss := openMockSvrs(3)
	cfg := &Config{
		Hosts:          collectHosts(tss),
		HostRetrys:     2,
		RetryTimeoutMs: 20,
	}
	c := New(cfg)

	data := []byte("helloworld")
	r := bytes.NewReader(data)
	ret := make(map[string]string)
	ctx := context.Background()

	svrs[0].Code = 570
	svrs[2].Code = 570
	svrs[1].Wait = &sync.WaitGroup{}
	svrs[1].Wait.Add(1)
	go func() {
		time.Sleep(500e6)
		svrs[1].Wait.Done()
	}()

	err := c.CallWith(ctx, &ret, "/200", "", r, r.Len())
	assert.NoError(t, err, "err: %#v", err)
	var count int
	for _, svr := range svrs {
		if datas, ok := svr.Bodys["200"]; ok {
			assert.Equal(t, 1, len(datas))
			assert.Equal(t, data, datas[0])
			count++
		}
	}
	assert.Equal(t, 1, count, "count: %v", count)

	closeServers(tss)
}
