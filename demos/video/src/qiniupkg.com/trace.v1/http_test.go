package trace_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"

	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/mockhttp.v2"
	"github.com/stretchr/testify/assert"
	"qiniupkg.com/qiniutest/httptest.v1"
	. "qiniupkg.com/trace.v1"
)

var testPath = "/tmp/trace-test/collector-dir"

type testService struct{}

func (p *testService) GetApi1(env *rpcutil.Env) error {
	return nil
}

func TestDefaultHTTPHandler(t *testing.T) {

	os.RemoveAll(testPath)

	clc, err := NewFileCollector(
		&FileCollectorConfig{
			LogDir:    testPath,
			ChunkBits: 10,
		})
	assert.Equal(t, nil, err, "new collector error")
	defer clc.Close()

	TracerEnable(
		SetService("papa-unit-test"),
		SetSampler(DummyTrueSampler),
		SetCollector(clc),
	)
	r := restrpc.Router{
		Mux: servestk.New(restrpc.NewServeMux(), HTTPHandler),
	}
	tr := mockhttp.NewTransport()
	tr.ListenAndServe("qiniu.trace.test", r.Register(&testService{}))

	ctx := httptest.New(t)
	ctx.SetTransport(tr)

	ctx.Exec(`
		get http://qiniu.trace.test/api1
		json '{}'
		ret 200
	`)
	err = DefaultTracer.Close()
	assert.Equal(t, nil, err, "close tracer error")

	b, err := ioutil.ReadFile(testPath + "/0")
	assert.Equal(t, nil, err, "open trace log error")

	span := &Span{}
	err = json.Unmarshal(b, span)
	assert.Equal(t, nil, err, "unmarshal trace log error")
	assert.Equal(t, span.IsRoot(), true, "span id info wrong")
}

func TestServeMux(t *testing.T) {

	os.RemoveAll(testPath)

	clc, err := NewFileCollector(
		&FileCollectorConfig{
			LogDir:    testPath,
			ChunkBits: 10,
		})
	assert.Equal(t, nil, err, "new collector error")
	defer clc.Close()

	TracerEnable(
		SetService("papa-unit-test"),
		SetSampler(DummyTrueSampler),
		SetCollector(clc),
	)
	r := restrpc.Router{
		Mux: NewServeMuxWith(restrpc.NewServeMux()),
	}
	tr := mockhttp.NewTransport()
	tr.ListenAndServe("qiniu.trace.test", r.Register(&testService{}))

	ctx := httptest.New(t)
	ctx.SetTransport(tr)

	ctx.Exec(`
		get http://qiniu.trace.test/api1
		json '{}'
		ret 200
	`)
	err = DefaultTracer.Close()
	assert.Equal(t, nil, err, "close tracer error")

	b, err := ioutil.ReadFile(testPath + "/0")
	assert.Equal(t, nil, err, "open trace log error")

	span := &Span{}
	err = json.Unmarshal(b, span)
	assert.Equal(t, nil, err, "unmarshal trace log error")
	assert.Equal(t, span.IsRoot(), true, "span id info wrong")
	assert.Equal(t, "Get/Api1", span.SpanName, "span not right")
}

func TestNewHTTPHandler(t *testing.T) {

	os.RemoveAll(testPath)
	clc, err := NewFileCollector(
		&FileCollectorConfig{
			LogDir:    testPath,
			ChunkBits: 10,
		})
	assert.Equal(t, nil, err, "new collector error")
	defer clc.Close()

	tracer := NewTracer(
		SetService("test-program"),
		SetSampler(DummyTrueSampler),
		SetCollector(clc),
	).Enable()

	r := restrpc.Router{
		Mux: servestk.New(restrpc.NewServeMux(), NewHTTPHandler(tracer)),
	}
	tr := mockhttp.NewTransport()
	tr.ListenAndServe("qiniu.trace.test", r.Register(&testService{}))

	ctx := httptest.New(t)
	ctx.SetTransport(tr)

	ctx.Exec(`
		get http://qiniu.trace.test/api1
		json '{}'
		ret 200
	`)
	err = tracer.Close()
	assert.Equal(t, nil, err, "close tracer error")

	b, err := ioutil.ReadFile(testPath + "/0")
	assert.Equal(t, nil, err, "open trace log error")

	span := &Span{}
	err = json.Unmarshal(b, span)
	assert.Equal(t, nil, err, "unmarshal trace log error")
	assert.Equal(t, span.IsRoot(), true, "span id info wrong")
}

func TestHTTPHandlerWithServiceCollector(t *testing.T) {

	os.RemoveAll(testPath)
	DefaultCollectRoot = testPath

	serviceName := "papa-unit-test"

	TracerEnable(
		SetService(serviceName),
		SetSampler(DummyTrueSampler),
	)
	tracer := DefaultTracer

	r := restrpc.Router{
		Mux: servestk.New(restrpc.NewServeMux(), HTTPHandler),
	}
	tr := mockhttp.NewTransport()
	tr.ListenAndServe("qiniu.trace.test", r.Register(&testService{}))

	ctx := httptest.New(t)
	ctx.SetTransport(tr)

	ctx.Exec(`
		get http://qiniu.trace.test/api1
		json '{}'
		ret 200
	`)
	err := tracer.Close()
	assert.Equal(t, nil, err, "close tracer error")

	logPath := path.Join(testPath, serviceName+"."+strconv.Itoa(os.Getpid()), "0")
	b, err := ioutil.ReadFile(logPath)
	assert.Equal(t, nil, err, "open trace log error")

	fmt.Println("=>", logPath)

	span := &Span{}
	err = json.Unmarshal(b, span)
	assert.Equal(t, nil, err, "unmarshal trace log error")
	assert.Equal(t, span.IsRoot(), true, "span id info wrong")
}

func TestConcurrency(t *testing.T) {

	os.RemoveAll(testPath)

	TracerEnable(
		SetService("papa-unit-test"),
		SetCollector(DummyCollector),
	)
	r := restrpc.Router{
		Mux: servestk.New(restrpc.NewServeMux(), HTTPHandler),
	}
	tr := mockhttp.NewTransport()
	tr.ListenAndServe("qiniu.trace.test", r.Register(&testService{}))

	cli := &http.Client{
		Transport: tr,
	}

	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				resp, err := cli.Get("http://qiniu.trace.test/api1")
				assert.Equal(t, nil, err, "request error")
				assert.Equal(t, 200, resp.StatusCode, "request error")
				resp.Body.Close()
			}
			wg.Done()
		}()
	}
	wg.Wait()

	err := DefaultTracer.Close()
	assert.Equal(t, nil, err, "close tracer error")
}

type rcWithWriterTo struct {
	io.Reader
}

func (rcw *rcWithWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	return rcw.Reader.(io.WriterTo).WriteTo(w)
}

func (rcw *rcWithWriterTo) Close() error {
	return nil
}

func TestTraceReadCloser(t *testing.T) {

	b := make([]byte, 12)
	buf := bytes.NewBuffer(b)

	// rc without WriterTo
	var rc io.ReadCloser
	rc = ioutil.NopCloser(bytes.NewBufferString("usa-election"))

	trc := ReadCloserWithTrace(DummyRecorder, rc, "test.read.closer")
	n, err := io.Copy(buf, trc)
	assert.Equal(t, 12, n, "io.Copy fail")
	assert.Equal(t, nil, err, "io.Copy fail")

	// rc with WriterTo
	rc = &rcWithWriterTo{bytes.NewBufferString("usa-election")}
	trc = ReadCloserWithTrace(DummyRecorder, rc, "test.read.closer")
	n, err = io.Copy(buf, trc)
	assert.Equal(t, 12, n, "io.Copy fail")
	assert.Equal(t, nil, err, "io.Copy fail")
}
