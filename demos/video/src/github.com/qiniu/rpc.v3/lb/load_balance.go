// 这个包bug比较多，不建议用。
package lb

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/rpc.v3"
	"github.com/qiniu/xlog.v1"
)

var ErrServiceNotAvailable = errors.New("service not available")

const (
	DefaultTryTimes          = 2
	DefaultFailRetryInterval = 60 // 1min
)

// --------------------------------------------------------------------
// func ShouldRetry

func ShouldRetry(err error) bool {

	if _, ok := err.(rpc.RespError); ok {
		return false
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	return true
}

// --------------------------------------------------------------------
// type Client

type conn struct {
	host           string
	lastFailedTime int64
}

type Config struct {
	Http              *http.Client
	ShouldRetry       func(error) bool
	FailRetryInterval int64
	TryTimes          uint32
}

type Client struct {
	conns             []conn
	client            rpc.Client
	shouldRetry       func(error) bool
	failRetryInterval int64
	tryTimes          uint32
	current           uint32
}

var defaultCfg Config

func New(hosts []string, cfg *Config) (p *Client, err error) {

	if len(hosts) == 0 {
		return nil, ErrServiceNotAvailable
	}

	if cfg == nil {
		cfg = &defaultCfg
	}

	conns := make([]conn, len(hosts))
	for i, host := range hosts {
		conns[i].host = host
	}

	client := rpc.Client{cfg.Http}
	if client.Client == nil {
		client.Client = http.DefaultClient
	}

	p = &Client{conns: conns, tryTimes: cfg.TryTimes, client: client, shouldRetry: cfg.ShouldRetry, failRetryInterval: cfg.FailRetryInterval}
	if p.tryTimes == 0 {
		p.tryTimes = DefaultTryTimes
	}
	if p.failRetryInterval == 0 {
		p.failRetryInterval = DefaultFailRetryInterval
	}
	if p.shouldRetry == nil {
		p.shouldRetry = ShouldRetry
	}
	return
}

func (p *Client) pickConn(fromIdx uint32) (c *conn, pickIdx uint32, err error) {

	n := len(p.conns)
	index := int(fromIdx)
	for i := 0; i < n; i++ {
		index = (index + 1) % n
		lastFailedTime := atomic.LoadInt64(&p.conns[index].lastFailedTime)
		if lastFailedTime == 0 || time.Now().Unix()-lastFailedTime >= p.failRetryInterval {
			return &p.conns[index], uint32(index), nil
		}
	}
	return nil, 0, ErrServiceNotAvailable
}

func (p *Client) Do(
	ctx context.Context, req *http.Request) (resp *http.Response, err error) {

	reqURI := req.URL.RequestURI()
	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx

		req.URL, err = url.Parse(c.host + reqURI)
		if err != nil {
			return
		}
		req.Host = c.host
		if strings.HasPrefix(c.host, "http://") {
			req.Host = c.host[len("http://"):]
		} else if strings.HasPrefix(c.host, "https://") {
			req.Host = c.host[len("https://"):]
		}

		resp, err = p.client.Do(ctx, req)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) CallWith64(
	ctx context.Context, ret interface{}, method, path string, bodyType string, body io.Reader, bodyLength int64) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWith64(ctx, ret, method, c.host+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) CallWithForm(ctx context.Context,
	ret interface{}, method, path string, params map[string][]string) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWithForm(ctx, ret, method, c.host+path, params)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return

}

func (p *Client) CallWithJson(ctx context.Context,
	ret interface{}, method, path string, params interface{}) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWithJson(ctx, ret, method, c.host+path, params)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return

}

func (p *Client) CallWith(ctx context.Context, ret interface{},
	method, path string, bodyType string, body io.Reader, bodyLength int) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWith(ctx, ret, method, c.host+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) Call(ctx context.Context, ret interface{}, method, path string) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.Call(ctx, ret, method, c.host+path)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) DoRequest(ctx context.Context, method, path string) (resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			xl := xlog.FromContextSafe(ctx)
			xl.Errorf("load balance DoRequest trytimes :%d ,err : %s", i, err2)
			break
		}
		fromIdx = pickIdx
		resp, err = p.client.DoRequest(ctx, method, c.host+path)
		if err != nil && p.shouldRetry(err) {
			xl := xlog.FromContextSafe(ctx)
			xl.Errorf("load balance DoRequest trytimes :%d ,err : %s", i, err)
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) DoRequestWith(ctx context.Context, method, path, bodyType string,
	body io.Reader, bodyLength int) (resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		resp, err = p.client.DoRequestWith(ctx, method, c.host+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			// err surely not be rpc.ErrorInfo
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

// --------------------------------------------------------------------
