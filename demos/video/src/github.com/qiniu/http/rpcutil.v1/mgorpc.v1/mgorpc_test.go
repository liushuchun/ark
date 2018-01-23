package mgorpc_v1

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/context"

	"net/http/httptest"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutputLevel(0)
}

// ---------------------------------------------------------------------------

type CustomEnv struct {
	w   http.ResponseWriter
	req *http.Request
}

func (p *CustomEnv) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) error {
	p.w = *w
	p.req = req
	return nil
}

func (p *CustomEnv) CloseEnv() {
	httputil.Reply(p.w, 200, map[string]int{"DoFoo_bar3": 3})
}

// ---------------------------------------------------------------------------

type Service struct {
}

type FooBar struct {
	Foo int    `json:"foo"`
	Bar string `json:"bar"`
}

type timeoutError struct{}

func (e *timeoutError) Error() string { return "no reachable servers" }

type WatermarkArgs struct {
	Mode  int    `flag:"_" json:"mode"`
	Image string `flag:"image,base64" json:"image"`
}

func (r *Service) WspFoo_bar(ctx context.Context, req *FooBar) (interface{}, error) {
	return nil, &timeoutError{}
}

func (r *Service) WspFoo_bar2(ctx context.Context, req *FooBar) (interface{}, error) {
	panic("[MGO2_COPY_SESSION_FAILED] servers failed")
	return nil, nil
}

func (r *Service) WsFoo_bar3(req *FooBar, env *CustomEnv) {
}

func (r *Service) WsFoo_bar_() error {
	return httputil.NewError(403, "bad request")
}

func (r *Service) CmdWatermark_(args *WatermarkArgs) (interface{}, error) {
	return nil, &timeoutError{}
}

func (r *Service) CmdWatermark4_(ctx context.Context, args *WatermarkArgs) (interface{}, error) {
	panic("[MGO2_COPY_SESSION_FAILED] servers failed")
	return nil, nil
}

func (r *Service) CmdWatermark2_(ctx context.Context, args *WatermarkArgs, env wsrpc.Env) {
	io.WriteString(env.W, "CmdWatermark2: "+args.Image)
}

func (r *Service) CmdpWatermark3(ctx context.Context) (interface{}, error) {
	return nil, &timeoutError{}
}

// ---------------------------------------------------------------------------

type mockMux struct {
	*http.ServeMux
}

type w2 struct {
	w http.ResponseWriter
}

func (w *w2) Header() http.Header {
	return w.w.Header()
}

func (w *w2) Write(b []byte) (int, error) {
	return w.w.Write(b)
}
func (w *w2) WriteHeader(i int) {
	w.w.WriteHeader(i)
}

func (mux *mockMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	mux.ServeMux.ServeHTTP(&w2{w}, r)
}

func TestRoute(t *testing.T) {
	service := new(Service)
	mux0 := &mockMux{http.NewServeMux()}
	router := webroute.Router{Factory: wsrpc.Factory, Mux: mux0}
	mux := router.Register(service)
	s := httptest.NewServer(mux)
	defer s.Close()
	time.Sleep(.5e9)
	{
		var ret map[string]map[string]interface{}
		param := map[string][]string{
			"foo": {"1"},
			"bar": {"abc"},
		}
		err := rpc.DefaultClient.CallWithForm(nil, &ret, s.URL+"/foo-bar", param)
		assert.Error(t, err)
		assert.Equal(t, 560, httputil.DetectCode(err))
	}
	{
		var ret WatermarkArgs
		err := rpc.DefaultClient.Call(nil, &ret, s.URL+"/watermark/1/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw")
		assert.Error(t, err)
		assert.Equal(t, 560, httputil.DetectCode(err))
	}
	{
		resp, err := http.Get(s.URL + "/watermark2/1/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw==")
		if err != nil {
			t.Fatal("http.Get failed:", err)
		}
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		if string(b) != "CmdWatermark2: http://www.b1.qiniudn.com/images/logo-2.png" {
			t.Fatal("call /watermark2 failed:", string(b))
		}
	}
	{
		var ret map[string]int
		err := rpc.DefaultClient.Call(nil, &ret, s.URL+"/foo-bar3?foo=1&bar=abc")
		if err != nil {
			t.Fatal("call /foo-bar3 failed:", err)
		}
	}
	{
		param := map[string][]string{
			"foo": {"1"},
			"bar": {"abc"},
		}
		err := rpc.DefaultClient.CallWithForm(nil, nil, s.URL+"/foo-bar/", param)
		if err == nil {
			t.Fatal("call /foo-bar/ failed:", err)
		}
		if e, ok := err.(*rpc.ErrorInfo); ok {
			if e.Code != 403 || e.Err != "bad request" {
				t.Fatal("call /foo-bar/ failed:", e)
			}
		} else {
			t.Fatal("call /foo-bar/ failed:", err)
		}
	}
	{
		resp, err := http.Get(s.URL + "/foo-bar?foo=1&bar=abc")
		if err != nil {
			t.Fatal("make conn failed", err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatal("get /foo-bar should get 405, but", resp.StatusCode)
		}
	}
	{
		resp, err := http.Get(s.URL + "/watermark3")
		if err != nil {
			t.Fatal("make conn failed", err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatal("get /wartermark3 should get 405, but", resp.StatusCode)
		}
	}
}
