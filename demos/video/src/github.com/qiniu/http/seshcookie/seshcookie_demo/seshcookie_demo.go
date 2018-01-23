package main

import (
	"fmt"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/seshcookie"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"net/http"
)

// -----------------------------------------------------------

type loginArgs struct {
	User     string `json:"user"`
	Password string `json:"pwd"`
}

type Service struct {
	seshcookie.Manager
	users map[string]string
}

func New() *Service {

	p := &Service{
		users: map[string]string{"foo": "bar", "xsw": "test"},
	}
	p.InitSessions("_session", "", "dace63516dd9fe281b9493eeecf8db85")
	return p
}

func (p *Service) WsLogin(args *loginArgs, env *seshcookie.Env) (err error) {

	if p.users[args.User] != args.Password {
		return httputil.NewError(401, "bad auth")
	}
	env.Session["user"] = args.User
	return nil
}

func (p *Service) WsLogout(env *seshcookie.Env) (err error) {

	delete(env.Session, "user")
	return nil
}

func (p *Service) WsHello(env *seshcookie.Env) {

	if user, ok := env.Session["user"]; ok {
		fmt.Fprintln(env.W, "Hello", user.(string))
	} else {
		fmt.Fprintln(env.W, "Sorry, I don't know who are you")
	}
}

// -----------------------------------------------------------

func main() {

	log.SetOutputLevel(0)

	service := New()
	router := &webroute.Router{Factory: wsrpc.Factory}

	log.Println("Starting service at port 9999")
	err := http.ListenAndServe(":9999", router.Register(service))
	log.Fatal("ListenAndServe:", err)
}

// -----------------------------------------------------------
