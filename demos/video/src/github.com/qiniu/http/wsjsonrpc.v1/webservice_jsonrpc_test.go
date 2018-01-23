package wsjsonrpc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
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
	Foo   int    `json:"foo"`
	Bar   string `json:"bar"`
	Mode  int    `flag:"_" json:"mode"`
	Image string `flag:"image,base64" json:"image"`
}

func (r *Service) WsprpcFoo_bar(req *FooBar) (map[string]interface{}, error) {
	return map[string]interface{}{"DoFoo_bar": req}, nil
}

func (r *Service) WsprpcFoo_bar3(req *FooBar, env *CustomEnv) {
}

func (r *Service) WsprpcFoo_bar_() error {
	return httputil.NewError(403, "bad request")
}

func (r *Service) CmdprpcWatermark_(args *FooBar) (interface{}, error) {
	return args, nil
}

func (r *Service) CmdprpcWatermark3() (interface{}, error) {
	return "hello", nil
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPage: "+req.URL.String())
}

// ---------------------------------------------------------------------------

func TestRoute(t *testing.T) {

	service := new(Service)
	router := webroute.Router{Factory: Factory}
	router.Register(service)
	svr := httptest.NewServer(router.Mux)
	u := svr.URL

	{
		var ret map[string]map[string]interface{}
		param := map[string]interface{}{
			"mode":  2,
			"image": "aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw",
		}
		err := rpc.DefaultClient.CallWithJson(nil, &ret, u+"/foo-bar?foo=1&bar=abc", param)
		if err != nil {
			t.Fatal("call /foo-bar failed:", err)
		}
		fmt.Println(ret)
		fb := ret["DoFoo_bar"]
		if fb == nil || fb["bar"].(string) != "abc" || fb["foo"].(float64) != 1 || fb["mode"].(float64) != 2 || fb["image"].(string) != "aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw" {
			t.Fatal("call /foo-bar failed:", ret)
		}
	}
	{
		var ret FooBar
		param := map[string]interface{}{
			"foo": 2,
			"bar": "abc",
		}
		err := rpc.DefaultClient.CallWithJson(nil, &ret, u+"/watermark/1/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw", param)
		if err != nil || ret.Mode != 1 || ret.Image != "http://www.b1.qiniudn.com/images/logo-2.png" || ret.Foo != 2 || ret.Bar != "abc" {
			t.Fatal("call /watermark/ failed:", ret, err)
		}
		fmt.Println(ret)
	}
	{
		var ret map[string]int
		err := rpc.DefaultClient.Call(nil, &ret, u+"/foo-bar3?foo=1&bar=abc")
		if err != nil {
			t.Fatal("call /foo-bar3 failed:", err)
		}
		fmt.Println(ret)
		if ret["DoFoo_bar3"] != 3 {
			t.Fatal("call /foo-bar3 failed:", ret)
		}
	}
	{
		param := map[string]interface{}{
			"mode":  2,
			"image": "aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw",
		}
		err := rpc.DefaultClient.CallWithJson(nil, nil, u+"/foo-bar/?foo=1&bar=abc", param)
		if err == nil {
			t.Fatal("call /foo-bar failed:", err)
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
		resp, err := http.Get(u + "/page?a=1&b=2")
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
		resp, err := http.Get(u + "/foo-bar?foo=1&bar=abc")
		if err != nil {
			t.Fatal("make conn failed", err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatal("get /foo-bar should get 405, but", resp.StatusCode)
		}
	}
	{
		resp, err := http.Get(u + "/watermark3")
		if err != nil {
			t.Fatal("make conn failed", err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatal("get /wartermark3 should get 405, but", resp.StatusCode)
		}
	}
}

// ---------------------------------------------------------------------------
