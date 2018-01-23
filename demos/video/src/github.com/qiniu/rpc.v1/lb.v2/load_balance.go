// 建议用lb.v2.1
// 存在问题：在开启 fail retry interval 时，可能会导致服务端负载不均衡
package lb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	cc "github.com/qiniu/io"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

var ErrServiceNotAvailable = errors.New("service not available")

const (
	DefaultTryTimes          = 2
	DefaultFailRetryInterval = 10 // 10s
)

// --------------------------------------------------------------------
// func ShouldRetry
func ShouldRetry(code int, err error) bool {
	if code == 0 {
		return true
	}
	if err == nil {
		return false
	}
	if _, ok := err.(rpc.RespError); ok {
		return false
	}
	return true
}

// --------------------------------------------------------------------
// type Request
type Request struct {
	http.Request
	Body io.ReaderAt
}

func NewRequest(method, urlStr string, body io.ReaderAt) (*Request, error) {
	var r io.Reader
	if body != nil {
		r = &cc.Reader{body, 0}
	}
	httpreq, err := http.NewRequest(method, urlStr, r)
	if err != nil {
		return nil, err
	}
	req := &Request{*httpreq, body}
	return req, nil
}

// --------------------------------------------------------------------
// go >= 1.5 roundtrip 同一个 request （重试）会导致后续的请求返回 error request canceled.
func discardAndClose(r io.ReadCloser) error {
	io.Copy(ioutil.Discard, r)
	return r.Close()
}

// --------------------------------------------------------------------
// type simpleClient
type conn struct {
	host           string
	lastFailedTime int64
}

type simpleConfig struct {
	client            *rpc.Client
	shouldRetry       func(int, error) bool
	failRetryInterval int64
	tryTimes          uint32
}

type simpleClient struct {
	*simpleConfig
	conns   []conn
	current uint32
}

func newSimpleClient(hosts []string, cfg *simpleConfig) *simpleClient {
	conns := make([]conn, len(hosts))
	for i, host := range hosts {
		conns[i].host = host
	}
	return &simpleClient{conns: conns, simpleConfig: cfg}
}

func (p *simpleClient) pickConn(fromIdx uint32) (c *conn, pickIdx uint32, err error) {
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

func (p *simpleClient) DoWithHostRet(l rpc.Logger, req *Request) (host string, resp *http.Response, code int, err error) {
	xl := xlog.NewWith(l)
	reqURI := req.URL.RequestURI()
	err = ErrServiceNotAvailable
	fromIdx := atomic.AddUint32(&p.current, 1)
	httpreq := req.Request

	var body2 io.ReadCloser
	if req.Body != nil {
		body2 = ioutil.NopCloser(&cc.Reader{req.Body, 0})
	}

	for i := 0; i < int(p.tryTimes); i++ {
		httpreq.Body = body2
		c, pickIdx, err2 := p.pickConn(fromIdx)
		if err2 != nil {
			break
		}
		host = c.host
		fromIdx = pickIdx

		httpreq.URL, err = url.Parse(c.host + reqURI)
		if err != nil {
			return
		}
		if req.Host == "" {
			httpreq.Host = httpreq.URL.Host

			// rollback to raw host, such as c.host = "www.google.com"
			if httpreq.Host == "" {
				httpreq.Host = host
			}
		}

		resp, err = p.client.Do(xl, &httpreq)
		code = 0
		if resp != nil {
			code = resp.StatusCode
		}
		if p.shouldRetry(code, err) {
			xl.Warn("load_balance retry, times: ", i, "code: ", code, "err: ", err)
			atomic.StoreInt64(&c.lastFailedTime, time.Now().Unix())
			if req.Body != nil {
				body2 = ioutil.NopCloser(&cc.Reader{req.Body, 0})
			}
			if resp != nil {
				if i != int(p.tryTimes)-1 {
					discardAndClose(resp.Body)
				}
			}
			continue
		}

		atomic.StoreInt64(&c.lastFailedTime, 0)
		return
	}
	return
}

// --------------------------------------------------------------------
// type Client
type Config struct {
	Http              *http.Client
	FailoverHttp      *http.Client
	ShouldRetry       func(int, error) bool
	FailRetryInterval int64
	TryTimes          uint32
	FailoverTryTimes  uint32
	ShouldFailover    func(int, error) bool
}

var defaultCfg Config

type Client struct {
	client         *simpleClient
	failover       *simpleClient
	shouldFailover func(int, error) bool
}

func New(hosts []string, cfg *Config) (p *Client, err error) {
	return NewWithFailover(hosts, nil, cfg)
}

func NewWithFailover(hosts []string, failoverHosts []string, cfg *Config) (p *Client, err error) {
	if len(hosts) == 0 {
		return nil, ErrServiceNotAvailable
	}
	if cfg == nil {
		cfg = &defaultCfg
	}
	if cfg.ShouldRetry == nil {
		cfg.ShouldRetry = ShouldRetry
	}
	if cfg.ShouldFailover == nil {
		cfg.ShouldFailover = cfg.ShouldRetry
	}
	if cfg.TryTimes == 0 {
		cfg.TryTimes = DefaultTryTimes
	}
	if cfg.FailoverTryTimes == 0 {
		cfg.FailoverTryTimes = cfg.TryTimes
	}
	if cfg.FailRetryInterval == 0 {
		cfg.FailRetryInterval = DefaultFailRetryInterval
	}
	if cfg.Http == nil {
		cfg.Http = http.DefaultClient
	}
	if cfg.FailoverHttp == nil {
		cfg.FailoverHttp = cfg.Http
	}

	sconf := &simpleConfig{client: &rpc.Client{cfg.Http}, shouldRetry: cfg.ShouldRetry, tryTimes: cfg.TryTimes, failRetryInterval: cfg.FailRetryInterval}
	client := newSimpleClient(hosts, sconf)

	var failoverClient *simpleClient
	if len(failoverHosts) != 0 {
		sconf := &simpleConfig{client: &rpc.Client{cfg.FailoverHttp}, shouldRetry: cfg.ShouldRetry, tryTimes: cfg.FailoverTryTimes, failRetryInterval: cfg.FailRetryInterval}
		failoverClient = newSimpleClient(failoverHosts, sconf)
	}

	p = &Client{client: client, failover: failoverClient, shouldFailover: cfg.ShouldFailover}
	return p, nil
}

func (p *Client) DoWithHostRet(
	l rpc.Logger, req *Request) (host string, resp *http.Response, err error) {
	xl := xlog.NewWith(l)
	host, resp, code, err := p.client.DoWithHostRet(xl, req)
	if p.failover == nil || !p.shouldFailover(code, err) {
		return
	}
	if resp != nil {
		discardAndClose(resp.Body)
	}
	xl.Warn("try failover client")
	host, resp, _, err = p.failover.DoWithHostRet(xl, req)
	return
}

func (p *Client) Do(
	l rpc.Logger, req *Request) (resp *http.Response, err error) {
	_, resp, err = p.DoWithHostRet(l, req)
	return
}

func (p *Client) CallWith64(
	l rpc.Logger, ret interface{}, path string, bodyType string, body io.ReaderAt, bodyLength int64) (err error) {

	resp, err := p.PostWith64(l, path, bodyType, body, bodyLength)
	if err != nil {
		return
	}
	return rpc.CallRet(l, ret, resp)
}

func (p *Client) CallWithForm(l rpc.Logger,
	ret interface{}, path string, params map[string][]string) (err error) {

	resp, err := p.PostWithForm(l, path, params)
	if err != nil {
		return
	}
	return rpc.CallRet(l, ret, resp)
}

func (p *Client) CallWithJson(l rpc.Logger,
	ret interface{}, path string, params interface{}) (err error) {

	resp, err := p.PostWithJson(l, path, params)
	if err != nil {
		return
	}
	return rpc.CallRet(l, ret, resp)
}

func (p *Client) CallWith(
	l rpc.Logger, ret interface{}, path string, bodyType string, body io.ReaderAt, bodyLength int) (err error) {

	resp, err := p.PostWith(l, path, bodyType, body, bodyLength)
	if err != nil {
		return err
	}
	return rpc.CallRet(l, ret, resp)
}

func (p *Client) Call(
	l rpc.Logger, ret interface{}, path string) (err error) {

	resp, err := p.PostEx(l, path)
	if err != nil {
		return err
	}
	return rpc.CallRet(l, ret, resp)
}

func (p *Client) PostEx(l rpc.Logger, path string) (resp *http.Response, err error) {
	req, err := NewRequest("POST", path, nil)
	if err != nil {
		return
	}
	return p.Do(l, req)
}

func (p *Client) PostWith64(
	l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int64) (resp *http.Response, err error) {
	req, err := NewRequest("POST", path, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = bodyLength
	return p.Do(l, req)
}

func (p *Client) PostWith(
	l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int) (resp *http.Response, err error) {
	_, resp, err = p.PostWithHostRet(l, path, bodyType, body, bodyLength)
	return
}

func (p *Client) PostWithForm(
	l rpc.Logger, path string, params map[string][]string) (resp *http.Response, err error) {
	msg := url.Values(params).Encode()
	_, resp, err = p.PostWithHostRet(l, path, "application/x-www-form-urlencoded", strings.NewReader(msg), len(msg))
	return
}

func (p *Client) PostWithJson(
	l rpc.Logger, path string, params interface{}) (resp *http.Response, err error) {
	msg, err := json.Marshal(params)
	if err != nil {
		return
	}
	_, resp, err = p.PostWithHostRet(l, path, "application/json", bytes.NewReader(msg), len(msg))
	return
}

func (p *Client) PostWithHostRet(
	l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int) (host string, resp *http.Response, err error) {
	req, err := NewRequest("POST", path, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(bodyLength)
	return p.DoWithHostRet(l, req)
}
