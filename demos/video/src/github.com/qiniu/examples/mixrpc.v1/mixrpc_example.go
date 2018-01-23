package mixrpc_example

import (
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
)

type fooInfo struct {
	Foo string `json:"foo"`
	A   string `json:"a"'`
	B   string `json:"b"`
	Id  string `json:"id"`
}

// ---------------------------------------------------------------------------

type Config struct {
}

type Service struct {
	foos map[string]fooInfo
}

func New(cfg *Config) (p *Service, err error) {

	p = &Service{
		foos: make(map[string]fooInfo),
	}
	return
}

// ---------------------------------------------------------------------------
// restrpc

type fooBarArgs struct {
	CmdArgs []string
	A       string `json:"a"'`
	B       string `json:"b"`
}

type fooBarRet struct {
	Id string `json:"id"`
}

/*
POST /foo/<FooArg>/bar
JSON {a: <A>, b: <B>}
 RET 200
JSON {id: <FooId>}
*/
func (p *Service) PostFoo_Bar(args *fooBarArgs, env *rpcutil.Env) (ret fooBarRet, err error) {

	id := args.A + "." + args.B
	p.foos[id] = fooInfo{
		Foo: args.CmdArgs[0],
		A: args.A,
		B: args.B,
		Id: id,
	}
	return fooBarRet{Id: id}, nil
}

type reqArgs struct {
	CmdArgs []string
}

/*
 GET /foo/<FooId>
 RET 200
JSON {a: <A>, b: <B>, foo: <Foo>, id: <FooId>}
*/
func (p *Service) GetFoo_(args *reqArgs, env *rpcutil.Env) (ret fooInfo, err error) {

	id := args.CmdArgs[0]
	if foo, ok := p.foos[id]; ok {
		return foo, nil
	}
	err = httputil.NewError(404, "id not found")
	return
}

// ---------------------------------------------------------------------------
// json rpc

/*
POST /double
JSON <Val>
 RET 200
JSON <Result>
*/
func (r *Service) RpcDouble(v int) (int, error) {
	return v * 2, nil
}

/*
POST /doubles
JSON [<Val1>, ...]
 RET 200
JSON [<Result1>, ...]
*/
func (r *Service) RpcDoubles(vs []int) ([]int, error) {
	for i, v := range vs {
		vs[i] = v * 2
	}
	return vs, nil
}

// ---------------------------------------------------------------------------
// ws rpc

type watermarkArgs struct {
	Mode  int    `flag:"_" json:"mode"`
	Image string `flag:"image,base64" json:"image"`
}

/*
POST /watermark/<Mode>/image/<Base64EncodedImage>
 RET 200
JSON {mode: <Mode>, image: <Image>}
*/
func (r *Service) CmdWatermark_(args *watermarkArgs) (interface{}, error) {
	return args, nil
}

// ---------------------------------------------------------------------------

