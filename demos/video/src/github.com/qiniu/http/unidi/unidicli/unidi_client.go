package unidicli

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"io"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var ErrReadTimeout = errors.New("read timeout")
var ErrWriteTimeout = errors.New("write timeout")

// ---------------------------------------------------------------------------

func RegisterProtocol(scheme string, rt http.RoundTripper) {

	tr := http.DefaultTransport.(*http.Transport)
	tr.RegisterProtocol(scheme, rt)
}

// ---------------------------------------------------------------------------

type closeSignaler struct {
	io.Reader
	done chan bool
}

func (p *closeSignaler) Close() error {
	p.done <- true
	return nil
}

// ---------------------------------------------------------------------------

type request struct {
	id   string
	data []byte
}

type connection struct {
	w       http.ResponseWriter
	reqchan chan request
}

type Transport struct {
	reqDatas    map[string]*connection // host => connection
	reqMutex    sync.Mutex
	respWaiters map[string]chan closeSignaler // reqId => waiter
	respMutex   sync.Mutex
	timeout     time.Duration
	reqId       uint64
}

func NewTransport(timeout time.Duration) *Transport {

	p := &Transport{
		reqDatas:    make(map[string]*connection),
		respWaiters: make(map[string]chan closeSignaler),
		timeout:     timeout,
	}
	return p
}

func (p *Transport) getConnection(host string) *connection {

	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()

	if c, ok := p.reqDatas[host]; ok {
		return c
	}
	reqchan := make(chan request, 16)
	c := &connection{reqchan: reqchan}
	p.reqDatas[host] = c
	return c
}

func (p *Transport) getRespWaiter(reqId string) (waiter chan closeSignaler, ok bool) {

	p.respMutex.Lock()
	defer p.respMutex.Unlock()

	waiter, ok = p.respWaiters[reqId]
	return
}

func (p *Transport) acquireRespWaiter(host string) (waiter chan closeSignaler, reqId string) {

	waiter = make(chan closeSignaler, 1)
	reqId = strconv.FormatUint(atomic.AddUint64(&p.reqId, 1), 36)

	p.respMutex.Lock()
	defer p.respMutex.Unlock()

	p.respWaiters[reqId] = waiter
	return
}

func (p *Transport) releaseRespWaiter(reqId string) {

	p.respMutex.Lock()
	defer p.respMutex.Unlock()

	delete(p.respWaiters, reqId)
}

/*
1) send request

请求包:

	POST /
	Reqhost: <UnidiSvrHost>

返回包:

	200 OK
	Reqid: <RequestId>
	Content-Length: <ContentLength>

	<HttpRequestData>

2) get response

请求包：

	POST /
	Reqid: <RequestId>

	<HttpResponseData>

返回包:

	200 OK
*/
func (p *Transport) ServeHTTP(w http.ResponseWriter, req1 *http.Request) {

	host := req1.Header.Get("Reqhost")

	// send request

	if host != "" {
		c := p.getConnection(host)
		c.w = w
		req := <-c.reqchan
		if c.w != w {
			c.reqchan <- req
			log.Info("unidicli: outdated connection")
			return
		}
		h := w.Header()
		h.Set("Reqid", req.id)
		h.Set("Content-Length", strconv.Itoa(len(req.data)))
		_, err := w.Write(req.data)
		if err != nil {
			log.Warn("unidicli: sendRequest failed -", err)
		}
		return
	}

	reqId := req1.Header.Get("Reqid")

	// get response

	if reqId != "" {
		if waiter, ok := p.getRespWaiter(reqId); ok {
			done := make(chan bool)
			waiter <- closeSignaler{req1.Body, done}
			<-done
			httputil.ReplyWithCode(w, 200)
			return
		}
	}

	httputil.ReplyErr(w, 400, "invalid request")
}

func (p *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	buf := bytes.NewBuffer(nil)
	err = req.Write(buf)
	if err != nil {
		return
	}

	waiter, reqId := p.acquireRespWaiter(req.Host)
	defer p.releaseRespWaiter(reqId)

	c := p.getConnection(req.Host)
	select {
	case c.reqchan <- request{reqId, buf.Bytes()}:
	case <-time.After(p.timeout):
		return nil, ErrWriteTimeout
	}

	select {
	case cs := <-waiter:
		br := bufio.NewReader(cs.Reader)
		resp, err = http.ReadResponse(br, req)
		if err != nil {
			cs.done <- true
		} else {
			cs.Reader, resp.Body = resp.Body, &cs
		}
		return
	case <-time.After(p.timeout):
		return nil, ErrReadTimeout
	}
}

// ---------------------------------------------------------------------------
