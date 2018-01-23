// +build go1.7

package proxytransport

import (
	"net/http"
	"time"
)

func NewTransport(conf Config) (*transport, error) {

	if conf.HostExpireS <= 0 {
		conf.HostExpireS = 60
	}
	if conf.DetectIntervalS <= 0 {
		conf.DetectIntervalS = 30
	}
	if conf.DefaultTransport == nil {
		conf.DefaultTransport = http.DefaultTransport
	}
	if conf.DirectTestDialer == nil {
		conf.DirectTestDialer = DefaultTimeoutDialer
	}
	for _, sc := range conf.Socks5 {
		dialer, err := NewSocks5Dialer(sc)
		if err != nil {
			return nil, err
		}
		conf.Dialers = append(conf.Dialers, dialer)
	}

	transports := make([]http.RoundTripper, len(conf.Dialers))
	for i, dial := range conf.Dialers {
		transports[i] = &http.Transport{
			Dial:                  dial.Dial,
			MaxIdleConns:          conf.MaxIdleConnsPerProxy,
			MaxIdleConnsPerHost:   conf.MaxIdleConnsPerHostPerProxy,
			IdleConnTimeout:       time.Millisecond * time.Duration(conf.IdleConnTimeoutMs),
			ResponseHeaderTimeout: time.Millisecond * time.Duration(conf.ResponseHeaderTimeoutMs),
		}
	}
	t := &transport{
		Config:           conf,
		defaultTransport: conf.DefaultTransport,
		proxyTransports:  transports,

		hostSpeeds:     make(map[string]*speedInfo),
		closed:         make(chan struct{}, 1),
		hostExpire:     time.Second * time.Duration(conf.HostExpireS),
		detectInterval: time.Second * time.Duration(conf.DetectIntervalS),
	}
	go t.startDetect()
	return t, nil
}
