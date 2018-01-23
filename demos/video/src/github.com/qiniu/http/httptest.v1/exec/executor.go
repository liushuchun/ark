package exec

import (
	"errors"
	"reflect"
	"strings"

	"github.com/qiniu/http/httptest.v1"
	"github.com/qiniu/osl/cmdargs.v1"
	"github.com/qiniu/osl/cmdline.v1"
)

// ---------------------------------------------------------------------------

type IContext interface{
	GetRawCmd() string
}

type IExternalContext interface {
	FindCmd(ctx IContext, cmd string) reflect.Value
}

var (
	External IExternalContext
	ExternalSub IExternalContext
)

// ---------------------------------------------------------------------------

type Context struct {
	rawCmd  string
	current interface{}
	autoVarMgr
}

func New() *Context {

	return &Context{}
}

func (p *Context) Exec(ctx *httptest.Context, code string) {

	sctx := &subContext{
		ctx: ctx,
		parent: p,
	}
	sctx.parser = cmdline.NewParser()
	sctx.parser.ExecSub = sctx.execSubCmd

retry:
	code, err := p.parseAndExec(ctx, sctx, code)
	if err == nil {
		goto retry
	}
}

func (p *Context) GetRawCmd() string {

	return p.rawCmd
}

func (p *Context) findCmd(cmd string) (method reflect.Value) {

	v := reflect.ValueOf(p)
	method = v.MethodByName("Cmd_" + cmd)
	if method.IsValid() {
		return
	}

	if External == nil {
		return
	}
	return External.FindCmd(p, cmd)
}

func (p *Context) parseAndExec(
	ctx *httptest.Context, sctx *subContext, code string) (codeNext string, err error) {

	baseFrame := p.enterFrame()
	defer p.leaveFrame(ctx, baseFrame)

	cmd, codeNext, err := sctx.parser.ParseCode(code)
	if err != nil && err != cmdline.EOF {
		ctx.Fatal(err)
		return
	}
	if len(cmd) > 0 {
		//
		// p.Cmd_xxx(ctx *httptest.Context, cmd []string)
		method := p.findCmd(cmd[0])
		if !method.IsValid() {
			ctx.Fatal("command not found:", cmd[0])
			return
		}
		cmdLen := len(code) - len(codeNext)
		p.rawCmd = strings.Trim(code[:cmdLen], " \t\r\n")
		ctx.Log("====>", p.rawCmd)
		_, err = runCmd(ctx, method, cmd)
		if err != nil {
			ctx.Fatal(cmd, "-", err)
			return
		}
	}
	return
}

func runCmd(ctx *httptest.Context, method reflect.Value, cmd []string) (out []reflect.Value, err error) {

	mtype := method.Type()
	if mtype.NumIn() != 2 {
		err = errors.New("invalid method prototype: method input argument count != 2")
		return
	}

	argsType := mtype.In(1)
	args, err := cmdargs.Parse(ctx.Context, argsType, cmd)
	if err != nil {
		err = errors.New("ParseArgs failed: " + err.Error())
		return
	}

	in := []reflect.Value{
		reflect.ValueOf(ctx),
		args,
	}
	return method.Call(in), nil
}

// ---------------------------------------------------------------------------

