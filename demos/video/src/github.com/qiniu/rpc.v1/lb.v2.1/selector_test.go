package lb

import (
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"qbox.us/ratelimit"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/require"
	"github.com/stretchr/testify/assert"
)

func TestHostStruct(t *testing.T) {
	xl := xlog.NewDummy()
	rh := "http://localhost:9040"
	u, err := url.Parse(rh)
	require.NoError(t, err)
	rl := ratelimit.New(0, 0)
	h := &host{raw: rh, URL: u, rl: rl}

	//require.True(t, h.IsOK(10))
	_, isPunished := h.IsPunished(10)
	require.True(t, !isPunished)
	h.SetFail(xl)
	//require.False(t, h.IsOK(10))
	_, isPunished = h.IsPunished(10)
	require.False(t, !isPunished)
	//require.True(t, h.IsOK(-1))
	_, isPunished = h.IsPunished(-1)
	require.True(t, !isPunished)
	h.lastFailedTime = time.Now().Add(-10 * time.Second).Unix()
	//require.True(t, h.IsOK(10))
	_, isPunished = h.IsPunished(10)
	require.True(t, !isPunished)
	h.SetFail(xl)
	//require.False(t, h.IsOK(1))
	_, isPunished = h.IsPunished(1)
	require.False(t, !isPunished)
	time.Sleep(2 * time.Second)
	//require.True(t, h.IsOK(1))
	_, isPunished = h.IsPunished(1)
	require.True(t, !isPunished)
}

func TestRetrySelector(t *testing.T) {
	xl := xlog.NewDummy()
	var rs retrySelector
	rs.idx = 1
	rs.failRetryInterval = 10
	rl := ratelimit.New(0, 0)
	rs.hosts = []*host{
		&host{raw: "host0", rl: rl},
		&host{raw: "host1", rl: rl},
		&host{raw: "host2", rl: rl},
		&host{raw: "host3", rl: rl},
	}
	rs.hosts[3].SetFail(xl)
	stat := make(map[string]int)
	N := 30
	for i := 0; i < (len(rs.hosts)-2)*N; i++ {
		h := rs.Get(xl)
		stat[h.raw]++
	}
	for i, host := range rs.hosts {

		//if i == int(rs.idx) || !host.IsOK(rs.failRetryInterval) {
		_, isPunished := host.IsPunished(rs.failRetryInterval)
		if i == int(rs.idx) || isPunished {
			require.Equal(t, 0, stat[host.raw], "%+v %v", host, stat)
		} else {
			require.Equal(t, N, stat[host.raw], "%+v %v", host, stat)
		}
	}

	rs.retryHosts = nil
	require.NotNil(t, rs.Get(xl))

	rs.retryHosts = nil
	rs.hosts[0].SetFail(xl)
	rs.hosts[2].SetFail(xl)
	require.Nil(t, rs.Get(xl))
}

func TestSelector(t *testing.T) {
	xl := xlog.NewDummy()
	sel := newSelector([]string{"http://host0", "http://host1", "http://host2"}, 0, 10, false, 0, nil, 0, 0)

	req, err := http.NewRequest("GET", "http://www.qiniu.com", nil)
	require.NoError(t, err)

	_, ok := sel.GetReqHost(req)
	require.False(t, ok)
	ehost := &host{raw: "abcd"}
	sel.SetReqHost(req, ehost)
	h, ok := sel.GetReqHost(req)
	require.True(t, ok)
	require.Equal(t, ehost, h)
	sel.DelReqHost(req)
	_, ok = sel.GetReqHost(req)
	require.False(t, ok)

	h, rs := sel.Get(xl)
	require.Equal(t, "http://host1", h.raw)
	require.Equal(t, "host1", h.URL.Host)

	h.SetFail(xl)
	for i := 0; i < 10; i++ {
		rhost := rs.Get(xl)
		require.NotEqual(t, "http://host1", rhost.raw)
	}

	h, rs = sel.Get(xl)
	require.Equal(t, "http://host2", h.raw)

	for i := 0; i < 10; i++ {
		rhost := rs.Get(xl)
		require.NotEqual(t, "http://host1", rhost.raw)
		require.NotEqual(t, "http://host2", rhost.raw)
	}

	h, rs = sel.Get(xl)
	require.Equal(t, "http://host0", h.raw)

	h, rs = sel.Get(xl)
	require.Equal(t, "http://host2", h.raw)
}

func TestLogwithPunishReqid(t *testing.T) {
	xl := xlog.NewWith("firstReq")
	sel := newSelector([]string{"http://host0", "http://host1", "http://host2"}, 0, 10, false, 0, nil, 0, 0)
	h, rs := sel.Get(xl)
	punishReqid, isPunished := h.IsPunished(10)
	require.Empty(t, punishReqid)
	require.True(t, !isPunished)
	require.Equal(t, "http://host1", h.raw)
	require.Equal(t, "host1", h.URL.Host)

	h.SetFail(xl)
	rhost := rs.Get(xl)
	require.NotEqual(t, "http://host1", rhost.raw)

	rs.hosts[2].SetFail(xl)
	rhost = rs.Get(xl)
	require.NotEqual(t, "http://host1", rhost.raw)

	xl = xlog.NewWith("secondReq")
	h, rs = sel.Get(xl)
	h.SetFail(xl)
	rhost = rs.Get(xl)
	require.Nil(t, nil, rhost)

}
func TestSelectorDns(t *testing.T) {
	xl := xlog.NewDummy()
	lookupCount := int32(0)
	lookupHost := func(host string) ([]string, error) {
		return []string{host + "-a", host + "-b"}, nil
	}
	LookupHost := func(host string) (addrs []string, err error) {
		atomic.AddInt32(&lookupCount, 1)
		return lookupHost(host)
	}
	sel := newSelector([]string{"http://host0", "http://host1:80", "https://host2"}, 0, 10, true, 1, LookupHost, 0, 0)

	var h *host
	for i := 0; i < 3; i++ {
		h, _ = sel.Get(xl)
		require.Equal(t, "http://host0-b", h.raw)
		require.Equal(t, "host0-b", h.URL.Host)
		if i == 2 {
			h.SetFail(xl)
		}

		h, _ = sel.Get(xl)
		require.Equal(t, "http://host1-a:80", h.raw)
		require.Equal(t, "host1-a:80", h.URL.Host)

		h, _ = sel.Get(xl)
		require.Equal(t, "http://host1-b:80", h.raw)
		require.Equal(t, "host1-b:80", h.URL.Host)
		if i == 2 {
			h.SetFail(xl)
		}

		h, _ = sel.Get(xl)
		require.Equal(t, "https://host2", h.raw)
		require.Equal(t, "host2", h.URL.Host)

		h, _ = sel.Get(xl)
		require.Equal(t, "http://host0-a", h.raw)
		require.Equal(t, "host0-a", h.URL.Host)
	}
	assert.Equal(t, int32(2), lookupCount)

	lookupHost = func(host string) ([]string, error) {
		return []string{host + "-b", host + "-c"}, nil
	}
	err := sel.resolveDns()
	assert.NoError(t, err)
	sel.hostsLastUpdateTime = time.Now().UnixNano()

	for i := 0; i < 3; i++ {
		h, _ = sel.Get(xl)
		require.Equal(t, "http://host0-c", h.raw)
		require.Equal(t, "host0-c", h.URL.Host)

		h, _ = sel.Get(xl)
		require.Equal(t, "http://host1-c:80", h.raw)
		require.Equal(t, "host1-c:80", h.URL.Host)

		h, _ = sel.Get(xl)
		require.Equal(t, "https://host2", h.raw)
		require.Equal(t, "host2", h.URL.Host)
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(4), lookupCount)
}

func TestSelectorDnsHost(t *testing.T) {
	xl := xlog.NewDummy()
	lookupHost := func(host string) ([]string, error) {
		return []string{"192.168.1.2", "192.168.1.1"}, nil
	}
	sel := newSelector([]string{"http://host1:80"}, 0, 10, true, 1, lookupHost, 0, 0)
	h, _ := sel.Get(xl)
	require.Equal(t, "host1:80", h.host)
}

func TestLookupHost(t *testing.T) {
	xl := xlog.NewDummy()

	hosts := []string{"192.168.1.5", "192.168.1.6"}
	lookupHost := func(host string) ([]string, error) {
		return hosts, nil
	}

	sel := newSelector([]string{"http://host1:321"}, 0, 10, true, 1, lookupHost, 0, 0)
	h, _ := sel.Get(xl)
	require.True(t, "http://192.168.1.5:321" == h.URL.String() || "http://192.168.1.6:321" == h.URL.String())

	require.NoError(t, sel.resolveDns())

	h, _ = sel.Get(xl)
	require.True(t, "http://192.168.1.5:321" == h.URL.String() || "http://192.168.1.6:321" == h.URL.String())

}
