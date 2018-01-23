package trace

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

type collectorFunc func(*Span) error

func (c collectorFunc) Collect(span *Span) error { return c(span) }

func (c collectorFunc) Close() error { return nil }

func TestRecorder(t *testing.T) {
	span := NewSpanWith(&SpanID{
		spanID:  spanID{1, 2, 3},
		sampled: true,
	})

	calledCollect := 0
	c := collectorFunc(func(s *Span) error {
		calledCollect++
		if s.spanID != span.spanID {
			t.Errorf("Collect: got spanID arg %v, want %v", s.spanID, span.spanID)
		}
		if s.Mode != MODE_ASYNC {
			t.Errorf("Collect: got spanID arg %v, want %v", s.Mode, span.Mode)
		}
		return nil
	})

	// check logs
	r := NewRecorder(span, c).Async()

	r.Log("log")

	r.Kv("key", "value")
	r.Kv("key", "valuee")

	r.Tag("key1", "value1")
	r.Tag("key1", "value2")

	r.Prof("prof-msg", time.Now().Add(-time.Second), time.Now())
	r.ProfKv("kv", "prof-msg", time.Now().Add(-time.Second), time.Now())
	r.ProfTag("tag", "prof-msg", time.Now().Add(-time.Second), time.Now())

	evt := struct {
		A int       `trace:"a"`
		T time.Time `trace:"t"`
	}{
		A: 1,
		T: time.Now(),
	}
	r.FlattenKV("test", evt)

	if len(r.span.KV) != 4 {
		t.Error("kv annotation wrong")
	}
	if len(r.span.Tag) != 2 {
		t.Error("tag annotation wrong")
	}
	if len(r.span.TS) != 3 {
		t.Error("ts annotation wrong")
	}

	// test ContextToken()
	req, _ := http.NewRequest("GET", "url", nil)
	r.Inject(req)
	token := req.Header.Get(TraceHeaderKey)
	if token != r.ContextToken() {
		t.Errorf("got wrong token %s, want %s", token, r.ContextToken())
	}

	// test NameOnce()
	r.Name("name1")
	if r.span.SpanName != "name1" {
		t.Errorf("wrong span name")
	}
	r.NameOnce("name2")
	if r.span.SpanName != "name1" {
		t.Errorf("wrong span name")
	}
	r.Name("name3")
	if r.span.SpanName != "name3" {
		t.Errorf("wrong span name")
	}

	if calledCollect != 0 {
		t.Errorf("got calledCollect %d, want 0", calledCollect)
	}
	r.Finish()

	if calledCollect != 1 {
		t.Errorf("got calledCollect %d, want 1", calledCollect)
	}
}

func TestShadowRecorder(t *testing.T) {
	span := NewRootSpan()
	rec := NewRecorder(span, DummyCollector)
	rec1 := rec.Shadow()

	if rec1.ContextToken() != rec.ContextToken() {
		t.Error("got wrong shadow recorder")
	}
}

func TestChildRecorder(t *testing.T) {
	span := NewRootSpan()
	rec := NewRecorder(span, DummyCollector)
	rec1 := rec.Child()

	if rec1.span.Parent != rec.span.Span {
		t.Error("got wrong child recorder")
	}
}

func TestDummyRecorder(t *testing.T) {
	dr := DummyRecorder

	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10000; j++ {
				dr.Reference()
				dr.Child()
				if dr.ContextToken() != "" {
					t.Fatal("wrong token")
				}
				dr.Log("hello")
				dr.LogAt("hello", time.Now())
				dr.FlattenKV("http", nil)
				dr.Client()
				dr.Mode("mode")
				dr.Hostname("hostname")
				dr.Kv("key", "value")
				dr.Tag("key1", "value1")
				dr.Tag("key1", "value2")
				dr.Prof("prof-msg", time.Now().Add(-time.Second), time.Now())
				dr.ProfKv("kv", "prof-msg", time.Now().Add(-time.Second), time.Now())
				dr.ProfTag("tag", "prof-msg", time.Now().Add(-time.Second), time.Now())

				dr.Finish()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	dr.Finish()
}

func TestWriteShadowRecorder(t *testing.T) {
	rec := NewRecorder(NewRootSpan(), DummyCollector)
	rec.span.Sample()

	token := rec.ContextToken()

	wg := sync.WaitGroup{}
	wg.Add(100)

	evt := struct {
		A int       `trace:"a"`
		T time.Time `trace:"t"`
	}{
		A: 1,
		T: time.Now(),
	}

	for i := 0; i < 100; i++ {
		go func(dr *Recorder) {
			for j := 0; j < 1000; j++ {
				if dr.ContextToken() != token {
					t.Fatal("wrong token")
				}
				dr.Client()
				dr.Mode("mode")
				dr.Hostname("hostname")

				// ts annotation
				dr.Log("hello")
				dr.LogAt("hello", time.Now())
				dr.Prof("prof-msg", time.Now().Add(-time.Second), time.Now())

				// kv annotation
				dr.Kv("key", "value")
				dr.ProfKv("kv", "prof-msg", time.Now().Add(-time.Second), time.Now())

				// 1 kv + 1 ts
				dr.FlattenKV("test", evt)

				// tag annotation
				dr.Tag("key1", "value1")
				dr.Tag("key1", "value2")
				dr.ProfTag("tag", "prof-msg", time.Now().Add(-time.Second), time.Now())
			}
			if dr.span.Len() > MaxAnnotations {
				t.Error("wrong annotation count", dr.span.Len())
			}
			wg.Done()
			dr.Finish()
		}(rec.Shadow())
	}
	wg.Wait()
	rec.Finish()
}
