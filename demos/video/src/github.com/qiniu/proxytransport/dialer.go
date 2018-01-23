package proxytransport

import (
	"net"
	"time"

	"golang.org/x/net/proxy"
)

type defaultDialer struct {
	timeout time.Duration
}

func (d defaultDialer) Dial(network, addr string) (net.Conn, error) {

	return net.DialTimeout(network, addr, d.timeout)
}

func NewTimeoutDialer(t time.Duration) Dialer {
	return defaultDialer{
		timeout: t,
	}
}

var DefaultTimeoutDialer = NewTimeoutDialer(time.Second)

// ---------------------------------------------------------

type Socks5Config struct {
	Host          string `json:"host"`
	User          string `json:"user"`
	Pass          string `json:"pass"`
	DialTimeoutMs int    `json:"dial_timeout_ms"`
}

func NewSocks5Dialer(conf Socks5Config) (Dialer, error) {

	var auth *proxy.Auth
	if conf.User != "" {
		auth = &proxy.Auth{
			conf.User,
			conf.Pass,
		}
	}
	dialer := DefaultTimeoutDialer
	if conf.DialTimeoutMs > 0 {
		dialer = NewTimeoutDialer(time.Duration(conf.DialTimeoutMs) * time.Millisecond)
	}
	return proxy.SOCKS5("tcp", conf.Host, auth, dialer)
}
