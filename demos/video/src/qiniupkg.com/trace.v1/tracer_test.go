package trace

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeCollector struct {
	called int32
}

func (c *fakeCollector) Collect(s *Span) error {
	atomic.AddInt32(&c.called, 1)
	return nil
}

func (c *fakeCollector) Close() error {
	return nil
}

func TestFromHTTP(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r := FromHTTP(req)
	r.Finish()
	r1 := FromHTTP(req)
	r1.Finish()
	assert.Equal(t, r.span.spanID, r1.span.spanID, "recorder not match")
}

func TestTraceReference1(t *testing.T) {

	fc := &fakeCollector{}
	tracer := NewTracer(SetCollector(fc), SetSampler(DummyTrueSampler)).Enable()

	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r := tracer.FromHTTP(req)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				r := tracer.FromHTTP(req)
				r.Log("hello").Finish()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	r.Log("hello").Finish()
	assert.Equal(t, 1, fc.called, "collect called wrong")
}

func TestTraceReference2(t *testing.T) {

	fc := &fakeCollector{}
	tracer := NewTracer(SetCollector(fc), SetSampler(DummySampler)).Enable()

	r := tracer.FromHTTP(nil)
	assert.Equal(t, DummyRecorder, r, "recorder wrong")

	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r = tracer.FromHTTP(req)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				r := tracer.FromHTTP(req)
				r.Log("hello").Finish()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	r.Log("hello").Finish()
	assert.Equal(t, 0, fc.called, "collect called wrong")
}

func TestRefFromHTTP1(t *testing.T) {

	fc := &fakeCollector{}
	tracer := NewTracer(SetCollector(fc), SetSampler(DummyTrueSampler)).Enable()

	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r := tracer.FromHTTP(req)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				r := tracer.RefFromHTTP(req)
				r.Log("hello")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	r.Log("hello").Finish()
	assert.Equal(t, 1, fc.called, "collect called wrong")
}

func TestRecDeadLock(t *testing.T) {

	fc := &fakeCollector{}
	tracer := NewTracer(SetCollector(fc), SetSampler(DummyTrueSampler)).Enable()

	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r := tracer.FromHTTP(req)

	ch := make(chan bool, 1)
	oldf := r.onFinish
	r.OnFinish(func() {
		ch <- true
		time.Sleep(1e9)
		oldf()
	})
	go r.Finish()

	<-ch
	tracer.FromHTTP(req)

	req, _ = http.NewRequest("GET", "http://a.com", nil)
	tracer.FromHTTP(req)
}

func TestRefFromHTTP2(t *testing.T) {

	fc := &fakeCollector{}
	tracer := NewTracer(SetCollector(fc), SetSampler(DummySampler)).Enable()

	r := tracer.RefFromHTTP(nil)
	assert.Equal(t, DummyRecorder, r, "recorder wrong")

	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r = tracer.RefFromHTTP(req)
	assert.Equal(t, DummyRecorder, r, "recorder wrong")
	r.Log("hello").Finish()
	assert.Equal(t, 0, fc.called, "collect called wrong")
}

func TestFromContextToken(t *testing.T) {
	fc := &fakeCollector{}
	tracer := NewTracer(SetCollector(fc), SetSampler(DummySampler)).Enable()

	// normal recorder
	token := NewRootSpan().ContextToken()
	r := tracer.FromContextToken(token)
	assert.Equal(t, token, r.ContextToken(), "token not match")

	// dummy recorder
	token = DummyRecorder.ContextToken()
	r = tracer.FromContextToken(token)
	assert.NotEqual(t, r, DummyRecorder, "recorder not right")

	req, _ := http.NewRequest("GET", "http://a.com", nil)
	r = tracer.RefFromHTTP(req)
	assert.Equal(t, DummyRecorder, r, "recorder wrong")
	r.Log("hello").Finish()
	assert.Equal(t, 0, fc.called, "collect called wrong")
}
