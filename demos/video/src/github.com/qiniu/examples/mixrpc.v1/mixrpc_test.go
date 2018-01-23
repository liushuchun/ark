package mixrpc_example

import (
	"testing"

	"github.com/qiniu/http/httptest.v1"
	"github.com/qiniu/http/httptest.v1/exec"
	"github.com/qiniu/http/jsonrpc.v1"
	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/mockhttp.v2"
)

func init() {
	log.SetOutputLevel(1)
}

// ---------------------------------------------------------------------------

func TestServer2(t *testing.T) {

	cfg := &Config{}

	svr, err := New(cfg)
	if err != nil {
		t.Fatal("New service failed:", err)
	}

	transport := mockhttp.NewTransport()
	router1 := webroute.Router{
		Factory: jsonrpc.Factory.Union(wsrpc.Factory),
	}
	router := restrpc.Router{
		PatternPrefix: "/v1",
		Default: router1.Register(svr),
	}
	transport.ListenAndServe("foo.com", router.Register(svr))

	ectx := exec.New()

	ctx := httptest.New(t)
	ctx.SetTransport(transport)

	ctx.Exec(ectx,
	`
	#
	# test restrpc
	#

	post http://foo.com/v1/foo/foo123/bar
	json '{"a": "1", "b": "2"}'
	ret 200
	json '{"id": $(id)}'

	get http://foo.com/v1/foo/$(id)
	ret 200
	json '{"id": $(id), "foo": "foo123", "a": "1", "b": "2"}'

	get http://foo.com/v1/foo/1.3
	ret 404
	json '{
		"error": "id not found"
	}'

	match $(abcd) 4578
	post http://foo.com/v1/foo/|base64 $(abcd)|/bar
	form a=$(id)&b=3
	ret 200
	json '{
		"id": $(id2)
	}'

	get http://foo.com/v1/foo/$(id2)
	ret 200
	json '{"foo": $(foo), "a": $(id), "b": "3"}'

	#
	# test wsrpc
	#

	post http://foo.com/watermark/1/image/|base64 http://open.qiniudn.com/images/abc.png|
	ret 200
	json '{
		"mode": 1,
		"image": "http://open.qiniudn.com/images/abc.png"
	}'

	#
	# test jsonrpc
	#

	post http://foo.com/double
	json 123
	ret 200
	json 246

	post http://foo.com/doubles
	json '[123, 4, 7]'
	ret 200
	json '[246, 8, 14]'
	`)
}

// ---------------------------------------------------------------------------

