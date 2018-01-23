// 这个包bug比较多，不建议用。建议用lb.v2.1
package lb

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/qiniu/rpc.v1"
)

var ErrServiceNotAvailable = errors.New("service not available")

const (
	DefaultTryTimes          = 2
	DefaultFailRetryInterval = 10 // 10s
)

// --------------------------------------------------------------------
// func ShouldRetry

func ShouldRetry(err error) bool {

	if _, ok := err.(rpc.RespError); ok {
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
	l rpc.Logger, req *http.Request) (resp *http.Response, err error) {

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

		// adjust request host with current host
		req.Host = req.URL.Host

		// rollback to raw host, such as c.host = "www.google.com"
		if req.Host == "" {
			req.Host = c.host
		}

		resp, err = p.client.Do(l, req)
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
	l rpc.Logger, ret interface{}, path string, bodyType string, body io.Reader, bodyLength int64) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWith64(l, ret, c.host+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) CallWithForm(l rpc.Logger,
	ret interface{}, path string, params map[string][]string) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWithForm(l, ret, c.host+path, params)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return

}

func (p *Client) CallWithJson(l rpc.Logger,
	ret interface{}, path string, params interface{}) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWithJson(l, ret, c.host+path, params)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return

}

func (p *Client) CallWith(
	l rpc.Logger, ret interface{}, path string, bodyType string, body io.Reader, bodyLength int) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.CallWith(l, ret, c.host+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) Call(
	l rpc.Logger, ret interface{}, path string) (err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		err = p.client.Call(l, ret, c.host+path)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) PostEx(l rpc.Logger, path string) (resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		resp, err = p.client.PostEx(l, c.host+path)
		if err != nil && p.shouldRetry(err) {
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			continue
		}
		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

func (p *Client) PostWith(
	l rpc.Logger, path, bodyType string, body io.Reader, bodyLength int) (resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		resp, err = p.client.PostWith(l, c.host+path, bodyType, body, bodyLength)
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

func (p *Client) PostWithHostRet(
	l rpc.Logger, path, bodyType string, body io.Reader, bodyLength int) (host string, resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < int(p.tryTimes); i++ {
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		fromIdx = pickIdx
		host = c.host
		resp, err = p.client.PostWith(l, c.host+path, bodyType, body, bodyLength)
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
