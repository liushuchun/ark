package reqid

import (
	"net"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReqid(t *testing.T) {

	var wg sync.WaitGroup

	ids := make([]string, 10)
	for i := 0; i < len(ids); i++ {
		wg.Add(1)
		go func(i int) {
			ids[i] = Gen()
			wg.Done()
		}(i)
	}

	wg.Wait()

	sort.Strings(ids)

	ipStr := newNetIPv4(ip).String()
	for i, id := range ids {
		info, err := Parse(id)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, info.IP, ipStr, "%v", i)
		assert.Equal(t, info.Pid, uint32(pid), "%v", i)
		assert.True(t, time.Now().Unix()-info.Unix < 100, "%v", i)
		assert.Equal(t, info.Index, uint32(i)+1, "%v", i)
		if i > 0 {
			assert.NotEqual(t, ids[i-1], id, "%v", i)
		}
	}

	_, err := Parse(ids[0][1:])
	assert.Error(t, err)

	_, err = Parse(ids[0][4:])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
}

func TestNetIPv4(t *testing.T) {

	var cases = []struct {
		IP      string
		Private bool
	}{
		{IP: "8.8.8.8", Private: false},
		{IP: "10.0.0.0", Private: true},
		{IP: "10.1.2.3", Private: true},
		{IP: "10.8.0.1", Private: true},
		{IP: "10.255.255.254", Private: true},
		{IP: "11.255.255.254", Private: false},
		{IP: "172.15.1.1", Private: false},
		{IP: "172.16.1.1", Private: true},
		{IP: "172.18.19.20", Private: true},
		{IP: "172.31.255.254", Private: true},
		{IP: "172.32.255.254", Private: false},
		{IP: "183.136.141.249", Private: false},
		{IP: "101.71.70.125", Private: false},
		{IP: "192.167.0.111", Private: false},
		{IP: "192.168.0.1", Private: true},
		{IP: "192.168.1.2", Private: true},
		{IP: "192.168.255.254", Private: true},
		{IP: "192.169.255.254", Private: false},
	}
	for i, c := range cases {
		ip := netIPv4(net.ParseIP(c.IP).To4())
		assert.Equal(t, c.IP, ip.String(), "%v", i)
		assert.Equal(t, c.Private, ip.IsPrivate())
		uip := ip.Uint32()
		ip = newNetIPv4(uip)
		assert.Equal(t, c.IP, ip.String(), "%v", i)
		assert.Equal(t, c.Private, ip.IsPrivate())
	}
}
