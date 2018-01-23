package unidisvr

import (
	"bufio"
	"github.com/qiniu/log.v1"
	"io"
	"net/http"
	"time"
)

// ---------------------------------------------------------------------------

type responseWriter struct {
	baseUrl, reqId string
	header         http.Header
	pw             *io.PipeWriter
	lastErr        error
	code           int
	writeHeader    bool
}

func newResponseWriter(baseUrl, reqId string) *responseWriter {

	return &responseWriter{
		baseUrl:     baseUrl,
		reqId:       reqId,
		code:        200,
		header:      make(http.Header),
		writeHeader: true,
	}
}

func (p *responseWriter) Header() http.Header {

	return p.header
}

func (p *responseWriter) WriteHeader(code int) {

	p.code = code
}

func (p *responseWriter) Write(buf []byte) (int, error) {

	if p.writeHeader {
		p.writeHeader = false
		pr, pw := io.Pipe()
		p.pw = pw
		go func() {
			resp := &http.Response{
				StatusCode:    p.code,
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        p.header,
				ContentLength: -1,
				Body:          pr,
			}
			err := p.reply(resp)
			pr.CloseWithError(err)
		}()
	}
	n, err := p.pw.Write(buf)
	if err != nil {
		p.lastErr = err
	}
	return n, err
}

func (p *responseWriter) Done() error {

	if p.writeHeader {
		p.writeHeader = false
		resp := &http.Response{
			StatusCode: p.code,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     p.header,
		}
		return p.reply(resp)
	} else if p.pw != nil {
		p.pw.CloseWithError(p.lastErr)
		p.pw = nil
	}
	return p.lastErr
}

/*
	POST /
	Reqid: <RequestId>

	<HttpResponseData>
*/
func (p *responseWriter) reply(resp *http.Response) error {

	pr, pw := io.Pipe()

	go func() {
		err := resp.Write(pw)
		pw.CloseWithError(err)
	}()

	req, _ := http.NewRequest("POST", p.baseUrl, pr)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Reqid", p.reqId)

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
	} else {
		log.Warn("unidisvr: Reply failed -", err)
	}
	pr.CloseWithError(err)
	return err
}

// ---------------------------------------------------------------------------

type Server struct {
	Addr          string
	BaseUrl       string
	Handler       http.Handler
	Transport     http.RoundTripper
	RetryDuration time.Duration
}

/*
	200 OK
	Reqid: <RequestId>
	Content-Length: <ContentLength>

	<HttpRequestData>
*/
func (p *Server) handle(resp *http.Response) {

	defer resp.Body.Close()

	reqId := resp.Header.Get("Reqid")
	if reqId == "" {
		log.Warn("unidisvr: Reqid not found")
		return
	}

	b := bufio.NewReader(resp.Body)
	req, err := http.ReadRequest(b)
	if err != nil {
		log.Warn("unidisvr: ReadRequest failed -", p.Addr, p.BaseUrl, err)
		return
	}

	w := newResponseWriter(p.BaseUrl, reqId)
	p.Handler.ServeHTTP(w, req)
	err = w.Done()
	if err != nil {
		log.Warn("unidisvr: ServeHTTP failed -", p.Addr, p.BaseUrl, err)
	}
}

func (p *Server) ListenAndServe() (err error) {

	if p.Handler == nil {
		p.Handler = http.DefaultServeMux
	}

	host := p.Addr
	baseUrl := p.BaseUrl
	retryDur := p.RetryDuration
	if retryDur == 0 {
		retryDur = 1e9
	}

	c := &http.Client{
		Transport: p.Transport,
	}

	for {
		// accept
		req, _ := http.NewRequest("POST", baseUrl, nil)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Reqhost", host)
		resp, err := c.Do(req)
		if err != nil {
			log.Warn("unidisvr: Post failed -", baseUrl, err)
			time.Sleep(retryDur)
			continue
		}
		// and then process
		go p.handle(resp)
	}
	return nil
}

func ListenAndServe(host, baseUrl string, h http.Handler) (err error) {

	svr := &Server{
		Addr:    host,
		BaseUrl: baseUrl,
		Handler: h,
	}
	return svr.ListenAndServe()
}

// ---------------------------------------------------------------------------
