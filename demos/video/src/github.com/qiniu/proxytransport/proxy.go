package proxytransport

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DefaultDetectDialTimeout = 2 * time.Second
)

type Dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}

type Config struct {
	Socks5 []Socks5Config `json:"socks5"`

	DefaultTransport            http.RoundTripper `json:"-"`
	DirectTestDialer            Dialer            `json:"-"`
	Dialers                     []Dialer          `json:"-"`
	MaxIdleConnsPerHostPerProxy int               `json:"max_idle_conns_per_host_per_proxy"`
	MaxIdleConnsPerProxy        int               `json:"max_idle_conns_per_proxy"`
	IdleConnTimeoutMs           int               `json:"idle_conn_timeout_ms"`
	ResponseHeaderTimeoutMs     int               `json:"response_header_timeout_ms"`

	HostExpireS     int `json:"host_expire_s"`
	DetectIntervalS int `json:"detect_interval_s"`
}

type transport struct {
	Config
	defaultTransport http.RoundTripper
	proxyTransports  []http.RoundTripper
	hostSpeeds       map[string]*speedInfo
	hostLock         sync.RWMutex
	closed           chan struct{}

	hostExpire     time.Duration
	detectInterval time.Duration
}

func (t *transport) startDetect() {

	for {
		time.Sleep(t.detectInterval)
		select {
		case <-t.closed:
			return
		default:
		}

		t.hostLock.RLock()
		m := make(map[string]*speedInfo, len(t.hostSpeeds))
		for k, v := range t.hostSpeeds {
			m[k] = v
		}
		t.hostLock.RUnlock()

		for host, info := range m {

			t.hostLock.Lock()
			if info.lastAccessTime.Add(t.hostExpire).Before(time.Now()) {
				delete(t.hostSpeeds, host)
				t.hostLock.Unlock()
				log.Printf("host(%s) is expired.", host)
				continue
			}
			t.hostLock.Unlock()

			if info.lastDetectTime.Add(t.detectInterval).After(time.Now()) {
				continue
			}

			go info.touch(t.DirectTestDialer, host, t.Dialers)
		}
	}
}

func (t *transport) Stop() {
	close(t.closed)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {

	idx, tp := t.sel(req.URL.Host)
	res, err := tp.RoundTrip(req)
	if err != nil {
		t.markFailed(req.URL.Host, idx, err)
	}
	return res, err
}

func (t *transport) addNew(host string) {

	t.hostLock.Lock()
	defer t.hostLock.Unlock()
	info, ok := t.hostSpeeds[host]
	if ok {
		return
	}

	info = &speedInfo{
		directCost:     0,
		byProxyCosts:   make([]time.Duration, len(t.Dialers)),
		lastAccessTime: time.Now(),
		lastDetectTime: time.Now(),
	}

	t.hostSpeeds[host] = info
	go info.touch(t.DirectTestDialer, host, t.Dialers)
}

func (t *transport) markFailed(host string, idx int, err error) {

	log.Printf("connect to (%s) with proxy(%d) failed, err: %v", host, idx, err)

	t.hostLock.RLock()
	info, ok := t.hostSpeeds[host]
	t.hostLock.RUnlock()
	if !ok {
		return
	}

	info.markFailed(idx)
}

func (t *transport) sel(host string) (int, http.RoundTripper) {

	t.hostLock.RLock()
	info, ok := t.hostSpeeds[host]
	t.hostLock.RUnlock()
	if !ok {
		go t.addNew(host)
		return -1, t.defaultTransport
	}

	idx := info.getFastestIdx()
	if idx == -1 {
		return -1, t.defaultTransport
	}
	return idx, t.proxyTransports[idx]
}

type speedInfo struct {
	directCost   time.Duration
	byProxyCosts []time.Duration

	lock           sync.RWMutex
	lastAccessTime time.Time
	lastDetectTime time.Time
}

func (s *speedInfo) markFailed(idx int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx == -1 {
		s.directCost = 0
	} else {
		s.byProxyCosts[idx] = 0
	}
}

func (s *speedInfo) getFastestIdx() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	minCost := s.directCost
	if minCost == 0 {
		minCost = time.Duration(math.MaxInt64)
	}
	idx := -1
	for i, cost := range s.byProxyCosts {
		if cost != 0 && cost < minCost {
			minCost = cost
			idx = i
		}
	}

	s.lastAccessTime = time.Now()

	return idx
}

func (s *speedInfo) touch(dtDialer Dialer, host string, proxys []Dialer) {

	directCost := dialCost(host, dtDialer)
	var proxyCosts = make([]time.Duration, len(proxys))

	for i, dialer := range proxys {
		proxyCosts[i] = dialCost(host, dialer)
	}
	log.Println("costs:", host, directCost, proxyCosts)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.lastDetectTime = time.Now()
	s.directCost = directCost
	s.byProxyCosts = proxyCosts
}

func dialCost(host string, dialer Dialer) time.Duration {

	var runtoc = make(chan bool, 3)
	go func() {
		for _, kw := range []string{"dial", "write", "read"} {
			select {
			case <-runtoc:
			case <-time.After(time.Second * 30):
				log.Printf("timeout with (%s) at %s\n", host, kw)
				return
			}
		}
	}()

	if strings.Index(host, ":") == -1 {
		host = host + ":80"
	}
	conn, err := dialer.Dial("tcp", host)
	runtoc <- true
	if err != nil {
		log.Printf("dial (%s) failed, err: %v", host, err)
		return 0
	}
	defer conn.Close()

	err = conn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		log.Println("SetWriteDeadline failed:", err)
	}

	dataStart := time.Now()
	_, err = fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	runtoc <- true
	if err != nil {
		log.Println("write conn failed:", err.Error())
		return 0
	}

	err = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	if err != nil {
		log.Println("SetReadDeadline failed:", err)
	}
	b := make([]byte, 1)
	_, err = conn.Read(b)
	runtoc <- true
	if err != nil {
		log.Println("read err:", err)
		return 0
	}
	dataCost := time.Now().Sub(dataStart)

	return dataCost
}
