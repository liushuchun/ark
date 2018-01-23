package lbnsq

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qiniupkg.com/x/log.v7"
)

var (
	ErrNotEnoughProducers = errors.New("not enough producers")
)

type Client struct {
	sync.RWMutex
	Cfg           *Config
	nsqdAddrs     []string
	nsqdCli       *rpc.Client
	producersCnt  int
	nsqlookupdCli *lb.Client
}

func New(cfg *Config) (cli *Client, err error) {
	err = validateConfig(cfg)
	if err != nil {
		return
	}
	cli = &Client{Cfg: cfg}
	cli.nsqlookupdCli = lb.New(&lb.Config{
		Hosts:    cfg.NsqLookupdAddrs,
		TryTimes: uint32(len(cfg.NsqLookupdAddrs)),
	}, http.DefaultTransport)

	c := rpc.NewClientTimeout(time.Millisecond*time.Duration(cfg.DialTimeoutMs), 0)
	c.Timeout = time.Millisecond * time.Duration(cfg.ClientTimeoutMs)
	cli.nsqdCli = &c
	err = cli.updateNsqdAddrs()
	if err != nil {
		return
	}
	go func() {
		for {
			time.Sleep(time.Duration(cfg.RefreshNsqdIntervalMs) * time.Millisecond)
			err := cli.updateNsqdAddrs()
			if err != nil {
				log.Error("updateNsqdAddrs", err)
			}
		}
	}()
	return cli, nil
}

func (c *Client) updateNsqdAddrs() (err error) {
	var nodesRet NodesRet
	err = c.nsqlookupdCli.GetCall(nil, &nodesRet, "/nodes")
	if err != nil {
		return
	}
	pCnt := len(nodesRet.Nsqds)
	if pCnt < c.Cfg.ConcurrencyCount {
		return ErrNotEnoughProducers
	}
	var addrs []string
	for _, producer := range nodesRet.Nsqds {
		addrs = append(addrs, fmt.Sprintf("http://%s:%d", producer.BroadcastAddress, producer.HttpPort))
	}
	c.Lock()
	c.nsqdAddrs = addrs
	c.producersCnt = pCnt
	c.Unlock()
	return
}

func (c *Client) Publish(topic string, body []byte) (err error) {
	//写策略：指定并发数量写入，如果有失败的，继续重试下一个，直到成功数量达到指定数量。
	hosts := make([]string, c.producersCnt)
	c.RLock()
	copy(hosts, c.nsqdAddrs)
	c.RUnlock()
	ps := &Hosts{hosts}

	w := c.Cfg.ConcurrencyCount
	if ps.Len() < w {
		err = ErrNotEnoughProducers
		return
	}
	var okCount, doingCount int
	var errCh = make(chan error, ps.Len())
	for okCount != w {
		if doingCount < w-okCount && ps.Len() > 0 {
			h := ps.Get()
			go func(host string) {
				errCh <- c.pub(host, topic, body)
			}(h)
			doingCount++
			continue
		}
		if doingCount == 0 {
			break
		}
		err = <-errCh
		doingCount--
		if err != nil {
			log.Error(err)
			continue
		}
		okCount++
	}
	if okCount != w {
		err = errors.New("publish msg failed")
	}
	return nil
}
