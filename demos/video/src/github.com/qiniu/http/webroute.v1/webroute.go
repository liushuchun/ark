package webroute

import (
	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/log.v1"
	"net/http"
	"reflect"
)

// ---------------------------------------------------------------------------

type Mux interface {
	Handle(pattern string, handler http.Handler)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type Router struct {
	Factory       hfac.HandlerFactory
	PatternPrefix string
	Mux           Mux
	Style         byte
}

func (r *Router) ListenAndServe(addr string, rcvr interface{}) error {

	return http.ListenAndServe(addr, r.Register(rcvr))
}

func (r *Router) Register(rcvr interface{}) Mux {

	if r.Mux == nil {
		r.Mux = http.DefaultServeMux
	}

	mux := r.Mux
	sep := r.Style
	factory := r.Factory
	routePrefix := r.PatternPrefix

	if sep == 0 {
		sep = '-'
	}
	if factory == nil {
		factory = hfac.Factory
	}

	typ := reflect.TypeOf(rcvr)
	rcvr1 := reflect.ValueOf(rcvr)

	// Install the methods
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		prefix, handler, err := factory.Create(rcvr1, method)
		if err != nil {
			continue
		}
		pattern := routePrefix + patternOf(method.Name[len(prefix):], sep)
		mux.Handle(pattern, handler)
		log.Debug("Install", pattern, "=>", method.Name)
	}

	return mux
}

func patternOf(method string, sep byte) string {

	var c byte
	route := make([]byte, 0, len(method)+8)
	for i := 0; i < len(method); i++ {
		c = method[i]
		if sep != '/' && c >= 'A' && c <= 'Z' {
			route = append(route, '/')
			c += ('a' - 'A')
		} else if c == '_' {
			c = sep
		}
		route = append(route, c)
	}
	if c == sep {
		route[len(route)-1] = '/'
	}
	return string(route)
}

// ---------------------------------------------------------------------------

func Register(mux Mux, rcvr interface{}) {

	router := &Router{Mux: mux}
	router.Register(rcvr)
}

func ListenAndServe(addr string, rcvr interface{}) error {

	router := new(Router)
	return http.ListenAndServe(addr, router.Register(rcvr))
}

// ---------------------------------------------------------------------------
