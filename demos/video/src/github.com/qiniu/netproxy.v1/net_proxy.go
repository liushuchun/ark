package netproxy

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	gohttputil "net/http/httputil"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/http/httputil.v1"
)

var (
	ErrNoProto = errors.New("no protocol")
	ErrUnsupportedProto = errors.New("unsupported protocol")

	ErrServiceNotFound = httputil.NewError(404, "service not found")
	ErrHostNotFound = httputil.NewError(404, "host not found")
)

// ------------------------------------------------------------------------

func nilDirector(req *http.Request) {}

var theProxy = &gohttputil.ReverseProxy{Director: nilDirector}

func proxy(w http.ResponseWriter, req *http.Request, endpoint string) {

	req.URL.Scheme = "http"
	req.URL.Host = endpoint
	theProxy.ServeHTTP(w, req)
}

func SetProxyTransport(t http.RoundTripper) {

	theProxy.Transport = t
}

// --------------------------------------------------------------------

type ServiceFinder interface {
	NextEndpoint(service string) (endpoint string, err error)
}

type ProxyConf struct {
	Service string `json:"service"`
	As      string `json:"as"`
	Host    string `json:"host"`
}

// --------------------------------------------------------------------

type httpProxy struct {
	confs    map[string]*ProxyConf // Host => ProxyConf
	mutex    sync.RWMutex
	services ServiceFinder
}

func newHttpProxy(parent *Service) *httpProxy {

	return &httpProxy{
		confs: make(map[string]*ProxyConf),
		services: parent.services,
	}
}

func (p *httpProxy) addProxy(proxy *ProxyConf) {

	p.mutex.Lock()
	p.confs[proxy.Host] = proxy
	p.mutex.Unlock()
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	host := req.Host

	p.mutex.RLock()
	conf, ok := p.confs[host]
	p.mutex.RUnlock()

	if !ok {
		httputil.Error(w, ErrHostNotFound)
		return
	}

	endpoint, err := p.services.NextEndpoint(conf.Service)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	proxy(w, req, endpoint)
}

func (p *httpProxy) run(endpoint string) {

	err := http.ListenAndServe(endpoint, p)
	log.Fatal("ListenAndServe(http proxy) failed:", err)
}

// --------------------------------------------------------------------

type Service struct {
	services   ServiceFinder
	httpProxys map[string]*httpProxy // endpoint => httpProxy
	mutex      sync.Mutex
}

func New(services ServiceFinder) *Service {

	return &Service{
		services: services,
		httpProxys: make(map[string]*httpProxy),
	}
}

func (p *Service) addHttpProxy(endpoint string, conf *ProxyConf) (err error) {

	p.mutex.Lock()
	proxy, ok := p.httpProxys[endpoint]
	if !ok {
		proxy = newHttpProxy(p)
		p.httpProxys[endpoint] = proxy
	}
	p.mutex.Unlock()

	if !ok {
		go proxy.run(endpoint)
	}
	proxy.addProxy(conf)
	return
}

func (p *Service) AddProxy(proxy *ProxyConf) (err error) {

	proto, endpoint, err := parseAs(proxy.As)
	if err != nil {
		return
	}

	switch proto {
	case "http":
		p.addHttpProxy(endpoint, proxy)
	case "tcp":
		ln, err1 := net.Listen("tcp", endpoint)
		if err1 != nil {
			log.Error("AddProxy: net.Listen failed -", err1)
			return err1
		}
		log.Info("Run proxy of", proxy.Service, "@", endpoint)
		go p.runTcpProxy(proxy.Service, ln)
	default:
		return ErrUnsupportedProto
	}
	return
}

func (p *Service) runTcpProxy(service string, ln net.Listener) {

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("netproxy.runTcpProxy: Accept failed -", err)
			continue
		}
		go p.handleTcpConn(conn, service)
	}
}

func (p *Service) handleTcpConn(in net.Conn, service string) {

	defer in.Close()

	endpoint, err := p.services.NextEndpoint(service)
	if err != nil {
		log.Error("netproxy.handleTcpConn: services.NextEndpoint failed -", err)
		return
	}

	out, err := net.Dial("tcp", endpoint)
	if err != nil {
		log.Error("netproxy.handleTcpConn: net.Dail failed -", err)
		return
	}
	defer out.Close()

	proxyTCP(in.(*net.TCPConn), out.(*net.TCPConn))
}

// proxyTCP proxies data bi-directionally between in and out.
//
func proxyTCP(in, out *net.TCPConn) {

	var wg sync.WaitGroup
	wg.Add(2)
	go copyBytes("from backend", in, out, &wg)
	go copyBytes("to backend", out, in, &wg)
	wg.Wait()
}

func copyBytes(direction string, dest, src *net.TCPConn, wg *sync.WaitGroup) {

	defer wg.Done()

	n, err := io.Copy(dest, src)
	if err != nil {
		log.Error("I/O error:", err, "n:", n, "direction:", direction)
	}
	dest.CloseWrite()
	src.CloseRead()
}

// <proto>://<endpoint>
//
func parseAs(as string) (proto, endpoint string, err error) {

	pos := strings.Index(as, ":")
	if pos <= 0 || !strings.HasPrefix(as[pos+1:], "//") {
		err = ErrNoProto
		return
	}

	proto = as[:pos]
	endpoint = as[pos+3:]
	return
}

// --------------------------------------------------------------------

