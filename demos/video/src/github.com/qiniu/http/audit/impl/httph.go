package impl

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/bytes/seekable"

	qbytes "github.com/qiniu/bytes"
	. "github.com/qiniu/http/audit/proto"
)

// ----------------------------------------------------------

type ResponseWriter struct {
	http.ResponseWriter
	body      *qbytes.Writer
	extra     M
	written   int64
	startT    int64
	mod       string
	code      int
	xlog      bool
	skip      bool
	noLogBody bool // Affect response with 2xx only.
}

func NewResponseWriter(
	w http.ResponseWriter,
	body *qbytes.Writer,
	code int,
	xlog bool,
	mod string,
	startT int64) *ResponseWriter {

	return &ResponseWriter{
		ResponseWriter: w,
		body:           body,
		code:           200,
		xlog:           xlog,
		mod:            mod,
		startT:         startT,
	}
}

const xlogKey = "X-Log"
const xwanKey = "X-Warn"
const maxXlogLen = 509 // 512 - len("...")

func (r *ResponseWriter) Write(buf []byte) (n int, err error) {
	if r.xlog {
		r.logDuration(r.code)
		fullXlog, trunced := r.xlogMerge()
		if trunced {
			defer func() {
				r.setXlog(fullXlog)
			}()
		}
		r.xlog = false
		if r.code/100 == 2 && r.noLogBody {
			r.skip = true
		}
	}
	n, err = r.ResponseWriter.Write(buf)
	r.written += int64(n)
	if n == len(buf) && !r.skip {
		n2, _ := r.body.Write(buf)
		if n2 == n {
			return
		}
	}
	r.skip = true
	return
}

func (r *ResponseWriter) GetBody() []byte {
	if r.skip {
		return nil
	}
	return r.body.Bytes()
}

func (r *ResponseWriter) GetWritten() int64 {
	return r.written
}

func (r *ResponseWriter) ExtraDisableBodyLog() {
	r.noLogBody = true
}

func (r *ResponseWriter) xlogMerge() (fullXlog string, trunc bool) {
	headers := r.Header()
	v, ok := headers[xlogKey]
	if !ok {
		return
	}

	defer func() {
		if len(fullXlog) > maxXlogLen {
			trunc = true
			headers[xlogKey] = []string{"..." + fullXlog[len(fullXlog)-maxXlogLen:]}
		}
	}()

	if len(v) <= 1 {
		fullXlog = v[0]
		return
	}
	fullXlog = strings.Join(v, ";")
	headers[xlogKey] = []string{fullXlog}
	return
}

func (r *ResponseWriter) setXlog(xlog string) {
	headers := r.Header()
	_, ok := headers[xlogKey]
	if !ok {
		return
	}
	headers[xlogKey] = []string{xlog}
}

//
// X-Log: xxx; MOD[:duration][/code]
//
func (r *ResponseWriter) WriteHeader(code int) {
	if r.xlog {
		r.logDuration(code)
		fullXlog, trunced := r.xlogMerge()
		if trunced {
			defer func() {
				r.setXlog(fullXlog)
			}()
		}
		r.xlog = false
		if r.code/100 == 2 && r.noLogBody {
			r.skip = true
		}
	}
	r.ResponseWriter.WriteHeader(code)
	r.code = code
}

func (r *ResponseWriter) ExtraWrite(key string, val interface{}) {
	if r.extra == nil {
		r.extra = make(M)
	}
	r.extra[key] = val
}

func (r *ResponseWriter) ExtraAddInt64(key string, val int64) {
	if r.extra == nil {
		r.extra = make(M)
	}
	if v, ok := r.extra[key]; ok {
		val += v.(int64)
	}
	r.extra[key] = val
}

func (r *ResponseWriter) ExtraAddString(key string, val string) {
	if r.extra == nil {
		r.extra = make(M)
	}
	var v []string
	if v1, ok := r.extra[key]; ok {
		v = v1.([]string)
	}
	r.extra[key] = append(v, val)
}

func (r *ResponseWriter) GetExtra() M {
	return r.extra
}

func (r *ResponseWriter) GetStatusCode() int {
	return r.code
}

func (r *ResponseWriter) logDuration(code int) {
	h := r.Header()
	dur := (time.Now().UnixNano() - r.startT) / 1e6
	xlog := r.mod
	if dur > 0 {
		xlog += ":" + strconv.FormatInt(dur, 10)
	}
	if code/100 != 2 {
		xlog += "/" + strconv.Itoa(code)
	}
	h[xlogKey] = append(h[xlogKey], xlog)
}

// ----------------------------------------------------------

func Info(w http.ResponseWriter, key string, val interface{}) {
	ew, ok := w.(extraWriter)
	if !ok {
		ew, ok = getExtraWriter(w)
	}
	if ok {
		ew.ExtraWrite(key, val)
	}
}

func AddInt64(w http.ResponseWriter, key string, val int64) {
	ew, ok := w.(extraInt64Adder)
	if !ok {
		ew, ok = getExtraInt64Adder(w)
	}
	if ok {
		ew.ExtraAddInt64(key, val)
	}
}

func Xwarn(w http.ResponseWriter, val string) {
	ew, ok := w.(extraStringAdder)
	if !ok {
		ew, ok = getExtraStringAdder(w)
	}
	if ok {
		ew.ExtraAddString(xwanKey, val)
	}
}

func DisableBodyLog(w http.ResponseWriter) {
	w1, ok := w.(extraBodyLogDisabler)
	if !ok {
		w1, ok = getExtraBodyLogDisabler(w)
	}
	if ok {
		w1.ExtraDisableBodyLog()
	}
}

// ----------------------------------------------------------

type Decoder interface {
	DecodeRequest(req *http.Request) (url_ string, header, params M)
	DecodeResponse(header http.Header, bodyThumb []byte, extra, params M) (resph M, body []byte)
}

type DecoderEx interface {
	DecodeRequestEx(req *http.Request) (api, url_ string, header, params M)
	DecodeResponse(header http.Header, bodyThumb []byte, extra, params M) (resph M, body []byte)
}

type BaseDecoder struct {
}

func set(h M, header http.Header, key string) {
	if v, ok := header[key]; ok {
		h[key] = v[0]
	}
}

func ip(addr string) string {
	pos := strings.Index(addr, ":")
	if pos < 0 {
		return addr
	}
	return addr[:pos]
}

func queryToJson(m map[string][]string) (h M, err error) {

	h = make(M)
	for k, v := range m {
		if len(v) == 1 {
			h[k] = v[0]
		} else {
			h[k] = v
		}
	}
	return
}

func (r BaseDecoder) DecodeRequest(req *http.Request) (url_ string, h, params M) {

	h = M{"IP": ip(req.RemoteAddr), "Host": req.Host}
	ct, ok := req.Header["Content-Type"]
	if ok {
		h["Content-Type"] = ct[0]
	}
	if req.URL.RawQuery != "" {
		h["RawQuery"] = req.URL.RawQuery
	}

	set(h, req.Header, "User-Agent")
	set(h, req.Header, "Range")
	set(h, req.Header, "Refer")
	set(h, req.Header, "Content-Length")
	set(h, req.Header, "If-None-Match")
	set(h, req.Header, "If-Modified-Since")
	set(h, req.Header, "X-Real-Ip")
	set(h, req.Header, "X-Forwarded-For")
	set(h, req.Header, "X-Scheme")
	set(h, req.Header, "X-Remote-Ip")
	set(h, req.Header, "X-Reqid")
	set(h, req.Header, "X-Id")
	set(h, req.Header, "X-From-Cdn")
	set(h, req.Header, "X-Tencent-Ua")
	set(h, req.Header, "Cdn-Src-Ip")
	set(h, req.Header, "Cdn-Scheme")
	set(h, req.Header, "X-Upload-Encoding")
	set(h, req.Header, "Accept-Encoding")

	// 记录非七牛 CDN 客户请求来源 CDN，方便排查问题
	// 网宿: wangsu, 蓝汛: ChinaCache, 帝联: dnion, 浩瀚: Power-By-HaoHan, 华云: 51CDN, 同兴: TXCDN
	set(h, req.Header, "Cdn")

	url_ = req.URL.Path
	if ok {
		switch ct[0] {
		case "application/x-www-form-urlencoded":
			seekable, err := seekable.New(req)
			if err == nil {
				req.ParseForm()
				params, _ = queryToJson(req.Form)
				seekable.SeekToBegin()
			}
		}
	}
	if params == nil {
		params = make(M)
	}
	return
}

func (r BaseDecoder) DecodeRequestEx(req *http.Request) (api, url_ string, h, params M) {
	url_, h, params = r.DecodeRequest(req)
	api = "/"
	return
}

func (r BaseDecoder) DecodeResponse(header http.Header, bodyThumb []byte, h, params M) (resph M, body []byte) {

	if h == nil {
		h = make(M)
	}

	ct, ok := header["Content-Type"]
	if ok {
		h["Content-Type"] = ct[0]
	}
	if xlog, ok := header["X-Log"]; ok {
		h["X-Log"] = xlog
	}
	set(h, header, "X-Reqid")
	set(h, header, "X-Id")
	set(h, header, "Content-Length")
	set(h, header, "Content-Encoding")

	if ok && ct[0] == "application/json" && header.Get("Content-Encoding") != "gzip" {
		if -1 == bytes.IndexAny(bodyThumb, "\n\r") {
			body = bodyThumb
		}
	}
	resph = h
	return
}

var DefaultDecoder BaseDecoder

// ----------------------------------------------------------
