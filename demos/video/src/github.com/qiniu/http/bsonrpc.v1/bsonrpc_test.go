package bsonrpc_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/log.v1"
	rpc "github.com/qiniu/rpc.v1/brpc"
)

func init() {
	log.SetOutputLevel(0)
}

// ---------------------------------------------------------------------------

type Service struct {
}

type FooArgs struct {
	Foo int    `json:"foo"`
	Bar string `json:"bar"`
}

type BarArgs struct {
	Foo int    `bson:"foo"`
	Bar string `bson:"bar"`
}

type BazArgs struct {
	M int    `flag:"_"`
	A int    `flag:"a"`
	B string `flag:"b"`
}

type DoubleArgs struct {
	V int `bson:"_"`
}

type DoublesArgs struct {
	VS []int `bson:"_"`
}

func (r *Service) WbrpcFoo(req *FooArgs, env bsonrpc.Env) (map[string]interface{}, error) {
	return map[string]interface{}{"Foo": req}, nil
}

func (r *Service) BbrpcBar(req *BarArgs, env bsonrpc.Env) (map[string]interface{}, error) {
	return map[string]interface{}{"Bar": req}, nil
}

func (r *Service) CmdbrpcBaz_(req *BazArgs, env bsonrpc.Env) (map[string]interface{}, error) {
	return map[string]interface{}{"M": req.M, "A": req.A, "B": req.B}, nil
}

func (r *Service) BbrpcDouble(req *DoubleArgs) (DoubleArgs, error) {
	req.V *= 2
	return *req, nil
}

func (r *Service) BbrpcDoubles(req *DoublesArgs) (*DoublesArgs, error) {
	for i, v := range req.VS {
		req.VS[i] = v * 2
	}
	return req, nil
}

// ---------------------------------------------------------------------------

func TestRoute(t *testing.T) {

	go func() {
		service := new(Service)
		router := webroute.Router{Factory: bsonrpc.Factory}
		t.Fatal(router.ListenAndServe(":3457", service))
	}()
	time.Sleep(.5e9)

	{
		var ret map[string]interface{}
		err := rpc.DefaultClient.CallWithForm(nil, &ret, "http://127.0.0.1:3457/foo", map[string][]string{
			"foo": {"1"},
			"bar": {"123"},
		})
		if err != nil {
			t.Fatal("call /foo failed:", err)
		}
		fmt.Println(ret)
		if ret["Foo"] == nil {
			t.Fatal("call /foo failed:", ret)
		}
	}
	{
		var ret map[string]interface{}
		err := rpc.DefaultClient.CallWithBson(nil, &ret, "http://127.0.0.1:3457/bar", &BarArgs{1, "123"})
		if err != nil {
			t.Fatal("call /bar failed:", err)
		}
		fmt.Println(ret)
		if ret["Bar"] == nil {
			t.Fatal("call /bar failed:", ret)
		}
	}
	{
		var ret struct {
			M int    `bson:"M"`
			A int    `bson:"A"`
			B string `bson:"B"`
		}
		err := rpc.DefaultClient.Call(nil, &ret, "http://127.0.0.1:3457/baz/0/a/1/b/val")
		if err != nil {
			t.Fatal("call /baz failed:", err)
		}
		fmt.Println(ret)
		if ret.M != 0 || ret.A != 1 || ret.B != "val" {
			t.Fatal("call /baz failed:", ret)
		}
	}
	{
		var ret DoubleArgs
		err := rpc.DefaultClient.CallWithBson(nil, &ret, "http://127.0.0.1:3457/double", &DoubleArgs{2})
		if err != nil || ret.V != 4 {
			t.Fatal("call /double failed:", ret, err)
		}
	}
	{
		var ret1 DoublesArgs
		err := rpc.DefaultClient.CallWithBson(nil, &ret1, "http://127.0.0.1:3457/doubles", &DoublesArgs{[]int{2, 3, 4}})
		ret := ret1.VS
		if err != nil || len(ret) != 3 || ret[0] != 4 || ret[1] != 6 || ret[2] != 8 {
			t.Fatal("call /doubles failed:", ret, err)
		}
	}
}

// ---------------------------------------------------------------------------
