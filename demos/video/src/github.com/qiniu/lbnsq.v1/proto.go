package lbnsq

import (
	"math/rand"
)

type Nsqd struct {
	RemoteAddress    string   `json:"remote_address"`
	Hostname         string   `json:"hostname"`
	BroadcastAddress string   `json:"broadcast_address"`
	TcpPort          int      `json:"tcp_port"`
	HttpPort         int      `json:"http_port"`
	Version          string   `json:"version"`
	Tombstones       []bool   `json:"tombstones"`
	Topics           []string `json:"topics"`
}

type NodesRet struct {
	Nsqds []Nsqd `json:"producers"`
}

type Hosts struct {
	hosts []string
}

func (p *Hosts) Len() int {
	return len(p.hosts)
}

func (p *Hosts) Get() (host string) {
	if p.Len() == 0 {
		panic("no host")
	}
	id := rand.Intn(p.Len())
	var p2 []string
	p2 = append(p2, p.hosts[:id]...)
	p2 = append(p2, p.hosts[id+1:]...)
	host, p.hosts = p.hosts[id], p2
	return
}
