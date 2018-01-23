package wsrpc_test

import (
	"fmt"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
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

type WatermarkArgs struct {
	Mode  int    `flag:"_" json:"mode"`
	Image string `flag:"image,base64" json:"image"`
}

func (r *Service) WspFoo_bar(req *FooBar) (map[string]interface{}, error) {
	return map[string]interface{}{"DoFoo_bar": req}, nil
}

func (r *Service) WsFoo_bar2(req *FooBar, env *wsrpc.Env) (map[string]interface{}, error) {
	return map[string]interface{}{"DoFoo_bar2": req}, nil
}

func (r *Service) WsFoo_bar3(req *FooBar, env *CustomEnv) {
}

func (r *Service) WsFoo_bar_() error {
	return httputil.NewError(403, "bad request")
}

func (r *Service) CmdWatermark_(args *WatermarkArgs) (interface{}, error) {
	return args, nil
}

func (r *Service) CmdWatermark2_(args *WatermarkArgs, env wsrpc.Env) {
	io.WriteString(env.W, "CmdWatermark2: "+args.Image)
}

func (r *Service) CmdpWatermark3() (interface{}, error) {
	return "hello", nil
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPage: "+req.URL.String())
}

// ---------------------------------------------------------------------------

func TestRoute(t *testing.T) {

	go func() {
		service := new(Service)
		router := webroute.Router{Factory: wsrpc.Factory}
		t.Fatal(router.ListenAndServe(":2458", service))
	}()
	time.Sleep(.5e9)

	{
		var ret map[string]map[string]interface{}
		param := map[string][]string{
			"foo": {"1"},
			"bar": {"abc"},
		}
		err := rpc.DefaultClient.CallWithForm(nil, &ret, "http://127.0.0.1:2458/foo-bar", param)
		if err != nil {
			t.Fatal("call /foo-bar failed:", err)
		}
		fmt.Println(ret)
		fb := ret["DoFoo_bar"]
		if fb == nil || fb["bar"].(string) != "abc" {
			t.Fatal("call /foo-bar failed:", ret)
		}
	}
	{
		var ret map[string]map[string]interface{}
		err := rpc.DefaultClient.Call(nil, &ret, "http://127.0.0.1:2458/foo-bar?foo=1&bar=abc")
		if err != nil {
			t.Fatal("call /foo-bar failed:", err)
		}
		fmt.Println(ret)
		fb := ret["DoFoo_bar"]
		if fb == nil || fb["bar"].(string) != "" {
			t.Fatal("call /foo-bar failed:", ret)
		}
	}
	{
		var ret WatermarkArgs
		err := rpc.DefaultClient.Call(nil, &ret, "http://127.0.0.1:2458/watermark/1/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw")
		if err != nil || ret.Mode != 1 || ret.Image != "http://www.b1.qiniudn.com/images/logo-2.png" {
			t.Fatal("call /watermark/ failed:", ret, err)
		}
		fmt.Println(ret)
	}
	{
		resp, err := http.Get("http://127.0.0.1:2458/watermark2/1/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw==")
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
		var ret map[string]map[string]interface{}
		err := rpc.DefaultClient.Call(nil, &ret, "http://127.0.0.1:2458/foo-bar2?foo=1&bar=abc")
		if err != nil {
			t.Fatal("call /foo-bar2 failed:", err)
		}
		fmt.Println(ret)
		fb := ret["DoFoo_bar2"]
		if fb == nil || fb["bar"].(string) != "abc" {
			t.Fatal("call /foo-bar2 failed:", ret)
		}
	}
	{
		var ret map[string]int
		err := rpc.DefaultClient.Call(nil, &ret, "http://127.0.0.1:2458/foo-bar3?foo=1&bar=abc")
		if err != nil {
			t.Fatal("call /foo-bar3 failed:", err)
		}
		fmt.Println(ret)
		if ret["DoFoo_bar3"] != 3 {
			t.Fatal("call /foo-bar3 failed:", ret)
		}
	}
	{
		param := map[string][]string{
			"foo": {"1"},
			"bar": {"abc"},
		}
		err := rpc.DefaultClient.CallWithForm(nil, nil, "http://127.0.0.1:2458/foo-bar/", param)
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
		resp, err := http.Get("http://127.0.0.1:2458/page?a=1&b=2")
		if err != nil {
			t.Fatal("http.Get failed:", err)
		}
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		if string(b) != "DoPage: /page?a=1&b=2" {
			t.Fatal("call /page failed:", string(b))
		}
	}

	{
		resp, err := http.Get("http://127.0.0.1:2458/foo-bar?foo=1&bar=abc")
		if err != nil {
			t.Fatal("make conn failed", err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatal("get /foo-bar should get 405, but", resp.StatusCode)
		}
	}
	{
		resp, err := http.Get("http://127.0.0.1:2458/watermark3")
		if err != nil {
			t.Fatal("make conn failed", err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatal("get /wartermark3 should get 405, but", resp.StatusCode)
		}
	}
}

// ---------------------------------------------------------------------------
