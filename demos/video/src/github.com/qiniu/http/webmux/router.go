package webmux

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
	Mux     Mux
	Factory hfac.HandlerFactory
}

func (r *Router) Register(rcvr interface{}, routes []string) Mux {

	if r.Mux == nil {
		r.Mux = http.DefaultServeMux
	}

	mux := r.Mux
	factory := r.Factory
	if factory == nil {
		factory = hfac.Factory
	}

	typ := reflect.TypeOf(rcvr)
	rcvr1 := reflect.ValueOf(rcvr)

	for i := 0; i+1 < len(routes); i += 2 {
		mname := routes[i+1]
		method, ok := typ.MethodByName(mname)
		if !ok {
			log.Warn("Install route failed: method not found -", mname)
			continue
		}
		_, handler, err := factory.Create(rcvr1, method)
		if err != nil {
			log.Warn("Install route failed:", mname, err)
			continue
		}
		mux.Handle(routes[i], handler)
		log.Debug("Install", routes[i], "=>", mname)
	}

	return mux
}

// ---------------------------------------------------------------------------
