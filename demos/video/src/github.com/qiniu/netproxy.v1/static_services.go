package netproxy

import (
	"sync/atomic"

	"github.com/qiniu/encoding.v2/jsonutil"
)

// --------------------------------------------------------------------

type serviceRoute struct {
	Hosts []string `json:"hosts"`
	index uint32
}

type StaticServices struct {
	serviceMap map[string]*serviceRoute // service => servieRoute
}

func NewStaticServices(conf string) (p *StaticServices, err error) {

	var cfg map[string]*serviceRoute
	err = jsonutil.Unmarshal(conf, &cfg)
	if err != nil {
		return
	}
	return &StaticServices{cfg}, nil
}

func (p *StaticServices) NextEndpoint(service string) (endpoint string, err error) {

	route, ok := p.serviceMap[service]
	if !ok {
		err = ErrServiceNotFound
		return
	}

	index := atomic.AddUint32(&route.index, 1)
	endpoint = route.Hosts[index % uint32(len(route.Hosts))]
	return
}

// --------------------------------------------------------------------

