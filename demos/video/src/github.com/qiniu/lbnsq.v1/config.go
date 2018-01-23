package lbnsq

import "qbox.us/errors"

type Config struct {
	NsqLookupdAddrs       []string `json:"nsq_lookupd_addrs"`
	RefreshNsqdIntervalMs int      `json:"refresh_nsqd_interval_ms"`
	ConcurrencyCount      int      `json:"concurrency_count"`
	ClientTimeoutMs       int      `json:"client_timeout_ms"`
	DialTimeoutMs         int      `json:"dial_timeout_ms"`
}

func validateConfig(cfg *Config) (err error) {
	if len(cfg.NsqLookupdAddrs) < 1 {
		return errors.New("empty nsq_lookupd_addrs")
	}
	if cfg.RefreshNsqdIntervalMs < 5000 {
		cfg.RefreshNsqdIntervalMs = 5000
	}
	if cfg.ConcurrencyCount < 1 {
		cfg.ConcurrencyCount = 1
	}
	if cfg.ClientTimeoutMs < 1 {
		cfg.ClientTimeoutMs = 1000
	}
	if cfg.DialTimeoutMs < 1 {
		cfg.DialTimeoutMs = 500
	}
	return
}
