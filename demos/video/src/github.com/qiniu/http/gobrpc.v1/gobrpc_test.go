package gobrpc_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/qiniu/http/gobrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/mockhttp.v2"

	"qiniupkg.com/http/httputil.v2"
	"qiniupkg.com/x/log.v7"
	"qiniupkg.com/x/rpc.v7"
	"qiniupkg.com/x/rpc.v7/gob"
)

var (
	gobClient = gob.DefaultClient
)

func init() {
	log.SetOutputLevel(0)

	gobClient.Transport = mockhttp.DefaultTransport
	http.DefaultClient.Transport = mockhttp.DefaultTransport
}

// ---------------------------------------------------------------------------

type Service struct {
}

type FooBar struct {
	Foo int
	Bar string
}

func init() {
	gob.RegisterName("FooBar", FooBar{})
}

func (r *Service) GobrpcFoo_bar(req *FooBar, env gobrpc.Env) (map[string]interface{}, error) {
	return map[string]interface{}{"DoFoo_bar": req}, nil
}

func (r *Service) GobrpcFoo_bar_() error {
	return httputil.NewError(403, "bad request")
}

func (r *Service) GobrpcDouble(v int) (int, error) {
	return v * 2, nil
}

func (r *Service) GobrpcDoubles(vs []int) ([]int, error) {
	for i, v := range vs {
		vs[i] = v * 2
	}
	return vs, nil
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPage: "+req.URL.String())
}

// ---------------------------------------------------------------------------

func TestRoute(t *testing.T) {

	service := new(Service)
	router := webroute.Router{Factory: gobrpc.Factory}
	mockhttp.ListenAndServe("127.0.0.1:2457", router.Register(service))

	{
		var ret map[string]interface{}
		err := gobClient.CallWithGob(nil, &ret, "POST", "http://127.0.0.1:2457/foo-bar", &FooBar{1, "123"})
		if err != nil {
			t.Fatal("call /foo-bar failed:", err)
		}
		fmt.Println(ret)
		if ret["DoFoo_bar"] == nil {
			t.Fatal("call /foo-bar failed:", ret)
		}
	}
	{
		err := gobClient.CallWithGob(nil, nil, "POST", "http://127.0.0.1:2457/foo-bar/", &FooBar{1, "123"})
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
		var ret int
		err := gobClient.CallWithGob(nil, &ret, "POST", "http://127.0.0.1:2457/double", 2)
		if err != nil || ret != 4 {
			t.Fatal("call /double failed:", ret, err)
		}
	}
	{
		var ret []int
		err := gobClient.CallWithGob(nil, &ret, "POST", "http://127.0.0.1:2457/doubles", []int{2, 3, 4})
		if err != nil || len(ret) != 3 || ret[0] != 4 || ret[1] != 6 || ret[2] != 8 {
			t.Fatal("call /doubles failed:", ret, err)
		}
	}
	{
		resp, err := http.Get("http://127.0.0.1:2457/page?a=1&b=2")
		if err != nil {
			t.Fatal("http.Get failed:", err)
		}
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		if string(b) != "DoPage: /page?a=1&b=2" {
			t.Fatal("call /page failed:", string(b))
		}
	}
}

// ---------------------------------------------------------------------------
