package httptest

import (
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/mockhttp.v2"
)

// ---------------------------------------------------------------------------

type mockTestingT struct {
	NilTestingT
	messages []string
	ok       bool
}

func (p *mockTestingT) Fatal(v ...interface{}) {

	log.Println(v...)
	if len(v) > 0 {
		if msg, ok := v[0].(string); ok {
			p.messages = append(p.messages, msg)
		}
	}
	p.ok = false
}

// ---------------------------------------------------------------------------

func init() {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		httputil.ReplyWith(w, 200, "application/text", []byte(req.URL.Path))
	})

	http.HandleFunc("/form", func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		httputil.Reply(w, 200, req.Form)
	})

	http.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
		h := w.Header()
		if ct, ok := req.Header["Content-Type"]; ok {
			h["Content-Type"] = ct
		}
		h.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
		w.WriteHeader(200)
		io.Copy(w, req.Body)
	})

	mockhttp.ListenAndServe("example.com", nil)
}

func Test_ContextDemo(t *testing.T) {

	ctx := New(t)
	ctx.SetTransport(mockhttp.DefaultTransport)

	ctx.Request("GET", "http://example.com/json").
		WithBody("json", `{"a": 1, "b": ["b1", "b2"]}`).
		Ret(200).
		WithBody("json", `{"a": 1}`)

	ctx.Request("GET", "http://example.com/form?a=1&b=b1&b=b2").
		Ret(200).
		WithBody("json", `{"a": ["1"], "b": ["b1", "b2"]}`)
}

// ---------------------------------------------------------------------------

type caseContext struct {
	method, url  string
	auth         interface{}
	reqHeader    http.Header
	reqBody      string
	reqBodyType  string
	code         int
	respHeader   http.Header
	respBody     string
	respBodyType string
	messages     []string
	ok           bool
}

func TestContext(t1 *testing.T) {

	cases := []caseContext{
		{
			method: "POST",
			url: "http://example.com/hello",
			auth: nil,
			reqHeader: http.Header{},
			reqBody: "",
			reqBodyType: "",
			code: 200,
			respHeader: http.Header{},
			respBody: "/hello",
			respBodyType: "application/text",
			messages: nil,
			ok: true,
		},
	}

	for _, c := range cases {
		t := &mockTestingT{ok: true}
		ctx := New(t)
		ctx.SetTransport(mockhttp.DefaultTransport)
		req := ctx.Request(c.method, c.url).WithAuth(c.auth)
		for k, v := range c.reqHeader {
			req.WithHeader(k, v...)
		}
		req.WithBody(c.reqBodyType, c.reqBody)
		resp := req.Ret(c.code)
		for k, v := range c.respHeader {
			resp.WithHeader(k, v...)
		}
		resp.WithBody(c.respBodyType, c.respBody)
		if !reflect.DeepEqual(t.messages, c.messages) || t.ok != c.ok {
			t1.Fatal("TestContext failed:", c, *t)
		}
	}
}

// ---------------------------------------------------------------------------

