package xlog

import (
	"bytes"
	"net/http"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/stretchr/testify/assert"

	"qiniupkg.com/trace.v1"
)

func TestXlog_Info(t *testing.T) {
	std := log.Std

	out := bytes.Buffer{}
	log.Std = log.New(&out, log.Std.Prefix(), log.Std.Flags())
	NewWith("RhQAAIfWo0-SNUwT").Info("test")
	outStr := out.String()
	assert.True(t, regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}.\d{6}$`).MatchString(outStr[:26]))
	assert.Equal(t, outStr[26:], " [RhQAAIfWo0-SNUwT][INFO] github.com/qiniu/xlog.v1/xlog_test.go:22: test\n")

	log.Std = std
}

// -----------------------------------------------------------------------------

type httpHeader http.Header

func (p httpHeader) ReqId() string {

	return p[reqidKey][0]
}

func (p httpHeader) Header() http.Header {

	return http.Header(p)
}

func TestNewWithHeader(t *testing.T) {

	reqid := "testnewwithheader"

	h := httpHeader(make(http.Header))
	h[logKey] = []string{"origin"}
	h[reqidKey] = []string{reqid}

	xlog := NewWith(h)
	xlog.Xput([]string{"append"})

	assert.Equal(t, h.ReqId(), reqid)
	assert.Equal(t, xlog.ReqId(), reqid)

	log := []string{"origin", "append"}
	assert.Equal(t, h[logKey], log)
	assert.Equal(t, xlog.Xget(), log)

	xlog.Xtag("v1")
	xlog.Xtag("v2")
	tags := []string{"v1;v2"}
	assert.Equal(t, h[tagKey], tags)
	xlog.Xtag("v3")
	assert.Equal(t, h[tagKey], []string{"v1;v2;v3"})
}

// -----------------------------------------------------------------------------

func TestGenReqId(t *testing.T) {

	reqId0 := GenReqId()

	for i, word := range []string{"hello", "world"} {
		SetGenReqId(func() string { return word })
		reqId := GenReqId()
		assert.Equal(t, word, reqId, "%v", i)
	}

	SetGenReqId(nil)
	reqId1 := GenReqId()
	assert.Equal(t, reqId0[:5], reqId1[:5])
}

func TestTraceRecorderConcurrent(t *testing.T) {

	trace.DefaultTracer.Enable()
	span := trace.NewRootSpan()
	span.Sample()

	trace.FromContextToken(span.ContextToken())

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "url", nil)
			trace.SetHTTPSpanID(span.SpanID, req, nil)

			for j := 0; j < 1000; j++ {
				l := NewWithReq(req)
				assert.Equal(t, span.ContextToken(), l.T().ContextToken(), "trace id not match")
				l1 := NewWith(l)
				assert.Equal(t, l, l1, "logger not match")

				l.T().Log("hello")
				l.T().Kv("key", "value")
				l.T().LogAt("hello", time.Now())
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestTraceRecorderConcurrent2(t *testing.T) {

	span := trace.DummyRecorder

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "url", nil)
			for j := 0; j < 1000; j++ {
				l := NewWithReq(req)
				assert.Equal(t, span.ContextToken(), l.T().ContextToken(), "trace id not match")
				l1 := NewWith(l)
				assert.Equal(t, l, l1, "logger not match")

				l.T().Log("hello")
				l.T().Kv("key", "value")
				l.T().LogAt("hello", time.Now())
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
