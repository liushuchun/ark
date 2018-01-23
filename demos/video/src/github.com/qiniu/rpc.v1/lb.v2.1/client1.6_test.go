// +build go1.6

package lb

import (
	"code.google.com/p/go.net/context"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v2/assert"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCancelRequest(t *testing.T) {
	xl := xlog.NewWith("CancelRequest")

	cfgsA := []*TestServerCfg{
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("A0"), RespHeaderTime: 3},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("A1"), RespHeaderTime: 3},
	}
	_, serversA, closer := startTestServers(t, cfgsA)

	proxys, _ := startProxys(2)

	var hostsA []string // 不通过代理访问
	for _, server := range serversA {
		hostsA = append(hostsA, server.URL)
	}
	var proxyHostsB []string
	for _, server := range proxys {
		proxyHostsB = append(proxyHostsB, server.URL)
	}

	tr := NewTransport(&TransportConfig{
		Proxys:        proxyHostsB,
		TryTimes:      uint32(10), // > 2x2
		ShouldReproxy: shouldReproxy,
	})

	cli := New(&Config{
		Hosts:       hostsA,
		TryTimes:    uint32(10), // > 2x2
		ShouldRetry: shouldRetry,
	}, tr)

	req, _ := NewRequest("POST", "/", strings.NewReader("testtesttest"))
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	req = req.WithContext(ctx)
	pc := atomic.LoadUint64(&proxycount)
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	_, err := cli.Do(xl, req)
	assert.Error(t, err)
	assert.Equal(t, 599, httputil.DetectCode(err))
	pc2 := atomic.LoadUint64(&proxycount)
	assert.Equal(t, uint64(1), pc2-pc)
	closer()

	cfgsA = []*TestServerCfg{
		&TestServerCfg{StatusCode: 575, ReturnBody: []byte("A0")},
		&TestServerCfg{StatusCode: 200, ReturnBody: []byte("A1")},
	}
	_, serversA, closer = startTestServers(t, cfgsA)
	defer closer()
	httpreq, _ := http.NewRequest("POST", serversA[0].URL, strings.NewReader("testtesttest"))
	resp, err := tr.RoundTrip(httpreq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 575)
	pc3 := atomic.LoadUint64(&proxycount)
	assert.Equal(t, uint64(1), pc3-pc2)

	httpreq, _ = http.NewRequest("POST", serversA[0].URL, nil)
	resp, err = tr.RoundTrip(httpreq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 575)
	pc4 := atomic.LoadUint64(&proxycount)
	assert.Equal(t, uint64(10), pc4-pc3)

	httpreq, _ = http.NewRequest("POST", serversA[0].URL, nil)
	reader := strings.NewReader("testtesttest")
	httpreq.Body = nopReadatCloser{reader, reader}
	httpreq.ContentLength = 12
	resp, err = tr.RoundTrip(httpreq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 575)
	pc5 := atomic.LoadUint64(&proxycount)
	assert.Equal(t, uint64(10), pc5-pc4)

}
