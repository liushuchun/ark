package transport

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
)

var (
	MaxBodyLength   int64 = 16 * 1024 * 1024
	ErrTooLargeBody       = errors.New("too large body")
)

type Transport struct {
	c *lb.Client
}

func NewTransport(c *lb.Client) *Transport {
	return &Transport{c}
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	xl := xlog.NewWithReq(req)
	req2, err := warpRequest(req)
	if err != nil {
		return
	}
	return t.c.Do(xl, req2)
}

func warpRequest(req *http.Request) (*lb.Request, error) {
	if req.ContentLength > MaxBodyLength {
		return nil, ErrTooLargeBody
	}
	body := io.ReaderAt(nil)
	if req.Body != nil {
		buf := new(bytes.Buffer)
		n, err := io.CopyN(buf, req.Body, MaxBodyLength+1)
		if n > MaxBodyLength {
			return nil, ErrTooLargeBody
		}
		if err != nil && err != io.EOF {
			return nil, err
		}
		body = bytes.NewReader(buf.Bytes())
	}
	return &lb.Request{Request: *req, Body: body}, nil
}
