package probe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/servestk.v1"

	qbytes "github.com/qiniu/bytes"
	. "github.com/qiniu/http/audit/impl"
	. "github.com/qiniu/http/audit/proto"
)

var HOST string = "localhost"

func init() {
	HOST, _ = os.Hostname()
}

// ----------------------------------------------------------

type ProbeWriter interface {
	Log(msg []byte) error
	Mark(measurement string,
		tim time.Time,
		tagK []string, tagV []string,
		fieldK []string, fieldV []interface{}) error
}

type Marker struct {
	w     ProbeWriter
	dec   DecoderEx
	event Event
	mod   string
	limit int
	xlog  bool
}

func New(mod string, w ProbeWriter, dec DecoderEx, limit int) *Marker {
	if dec == nil {
		dec = DefaultDecoder
	}
	return &Marker{w, dec, nil, mod, limit, true}
}

func NewEx(mod string, w ProbeWriter, dec DecoderEx, limit int, xlog bool) *Marker {
	if dec == nil {
		dec = DefaultDecoder
	}
	return &Marker{w, dec, nil, mod, limit, xlog}
}

func (r *Marker) SetEvent(event Event) {

	r.event = event
}

func (r *Marker) Handler(
	w http.ResponseWriter, req *http.Request, f func(http.ResponseWriter, *http.Request)) {

	api, url_, headerM, paramsM := r.dec.DecodeRequestEx(req)
	if url_ == "" { // skip
		servestk.SafeHandler(w, req, f)
		return
	}

	var header, params, resph []byte
	if len(headerM) != 0 {
		header, _ = json.Marshal(headerM)
	}
	if len(paramsM) != 0 {
		params, _ = json.Marshal(paramsM)
		if len(params) > 4096 {
			params, _ = json.Marshal(M{"discarded": len(params)})
		}
	}

	body := qbytes.NewWriter(make([]byte, r.limit))
	b := bytes.NewBuffer(nil)
	startT := time.Now()
	startTime := startT.UnixNano()
	w1 := NewResponseWriter(
		w,
		body,
		200,
		r.xlog,
		r.mod,
		startTime,
	)

	event := r.event
	if event == nil {
		servestk.SafeHandler(w1, req, f)
	} else {
		req1 := &Request{
			StartTime: startTime,
			Method:    req.Method,
			Mod:       r.mod,
			Path:      url_,
			Header:    headerM,
			Params:    paramsM,
		}
		id, err := event.OnStartReq(req1)
		if err != nil {
			httputil.Error(w1, err)
		} else {
			servestk.SafeHandler(w1, req, f)
			event.OnEndReq(id)
		}
	}

	startTime /= 100
	endTime := time.Now().UnixNano() / 100

	b.WriteString("REQ\t")
	b.WriteString(r.mod)
	b.WriteByte('\t')

	b.WriteString(strconv.FormatInt(startTime, 10))
	b.WriteByte('\t')
	b.WriteString(req.Method)
	b.WriteByte('\t')
	b.WriteString(url_)
	b.WriteByte('\t')
	b.Write(header)
	b.WriteByte('\t')
	b.Write(params)
	b.WriteByte('\t')

	resphM, respb := r.dec.DecodeResponse(w1.Header(), w1.GetBody(), w1.GetExtra(), paramsM)
	if len(resphM) != 0 {
		resph, _ = json.Marshal(resphM)
	}

	var code string = strconv.Itoa(w1.GetStatusCode())

	b.WriteString(code)
	b.WriteByte('\t')
	b.Write(resph)
	b.WriteByte('\t')
	b.Write(respb)
	b.WriteByte('\t')
	b.WriteString(strconv.FormatInt(w1.GetWritten(), 10))
	b.WriteByte('\t')
	b.WriteString(strconv.FormatInt(endTime-startTime, 10))

	err := r.w.Log(b.Bytes())
	if err != nil {
		errors.Info(err, "jsonlog.Handler: Log failed").Detail(err).Warn()
	}

	// qiniu_req,mod=xx,api=xx,host=xx,code=xx dur=xx
	r.w.Mark("qiniu_req", startT,
		[]string{"mod", "api", "host", "code"}, []string{r.mod, api, HOST, code},
		[]string{"dur"}, []interface{}{endTime - startTime})
}

// ----------------------------------------------------------
