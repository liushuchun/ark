package pipe

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------

type Writer struct { // 符合：http.ResonseWriter + CloseWithError
	pw          *io.PipeWriter
	header      http.Header
	done        chan<- bool
	code        int
	wroteHeader bool
}

func newWriter(pw *io.PipeWriter, done chan<- bool) *Writer {

	return &Writer{
		pw:     pw,
		header: make(http.Header),
		done:   done,
	}
}

func (rw *Writer) Header() http.Header {

	return rw.header
}

func (rw *Writer) Write(buf []byte) (int, error) {

	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}
	return rw.pw.Write(buf)
}

func (rw *Writer) WriteHeader(code int) {

	if !rw.wroteHeader {
		rw.code = code
		rw.wroteHeader = true
		rw.done <- true
	}
}

func (rw *Writer) CloseWithError(err error) error { // 确保 rw.WriteHeader 没被调用也仍然有 done 消息发出

	if !rw.wroteHeader {
		rw.done <- false
	}
	return rw.pw.CloseWithError(err)
}

// ---------------------------------------------------------------------------

var ErrBadContentLength = errors.New("bad Content-Length")

type Reader struct {
	pr   *io.PipeReader
	w    *Writer
	done <-chan bool
}

func newReader(pr *io.PipeReader, w *Writer, done <-chan bool) *Reader {

	return &Reader{
		pr:   pr,
		w:    w,
		done: done,
	}
}

func parseContentLength(cl string) (int64, error) {

	cl = strings.TrimSpace(cl)
	if cl == "" {
		return -1, nil
	}
	n, err := strconv.ParseInt(cl, 10, 64)
	if err != nil || n < 0 {
		return 0, ErrBadContentLength
	}
	return n, nil
}

func (rr *Reader) Get() (*http.Response, error) {

	<-rr.done

	header := rr.w.header
	ctxlen, err := parseContentLength(header.Get("Content-Length"))
	if err != nil {
		rr.pr.CloseWithError(err)
		return nil, err
	}

	return &http.Response{ // 只写了常用字段
		StatusCode:    rr.w.code,
		Header:        header,
		Body:          rr.pr,
		ContentLength: ctxlen,
	}, nil
}

// ---------------------------------------------------------------------------

/*

使用方式：

	import "github.com/qiniu/http/httputil/pipe"

	rr, rw := pipe.New()
	go func() {
		defer rw.CloseWithError(nil)
		fop(rw, ...)
	}()

	resp, err := rr.Get()
	...
*/
func New() (rr *Reader, rw *Writer) {

	done := make(chan bool, 1)
	pr, pw := io.Pipe()

	rw = newWriter(pw, done)
	rr = newReader(pr, rw, done)
	return
}

// ---------------------------------------------------------------------------
