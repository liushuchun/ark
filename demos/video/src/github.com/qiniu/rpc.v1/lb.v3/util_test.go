package lb

import (
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyExcept(t *testing.T) {

	ss := []string{"a", "b", "c", "d", "e"}
	idxs := []int{0, 1, 2, 3, 4}
	rets := [][]string{
		{"b", "c", "d", "e"},
		{"a", "c", "d", "e"},
		{"a", "b", "d", "e"},
		{"a", "b", "c", "e"},
		{"a", "b", "c", "d"},
	}
	for i, idx := range idxs {
		ret := copyExcept(ss, idx)
		assert.Equal(t, rets[i], ret, "%v", i)
	}
}

func TestRandomShrink(t *testing.T) {

	all := []string{"a", "b", "c", "d", "e"}
	ss := []string{"a", "b", "c", "d", "e"}
	s := ""
	all0 := []string{}
	for _ = range all {
		ss, s = randomShrink(ss)
		all0 = append(all0, s)
	}
	sort.StringSlice(all0).Sort()
	assert.Equal(t, all, all0)
}

func TestIndexRequest(t *testing.T) {

	rs := make([]*http.Request, 5)
	for i := range rs {
		rs[i] = new(http.Request)
	}
	assert.Equal(t, -1, indexRequest(rs[0:0], rs[0]))
	assert.Equal(t, 0, indexRequest(rs[0:1], rs[0]))
	assert.Equal(t, -1, indexRequest(rs[1:], rs[0]))
	assert.Equal(t, 0, indexRequest(rs[1:], rs[1]))
	assert.Equal(t, 1, indexRequest(rs[1:], rs[2]))
	assert.Equal(t, 2, indexRequest(rs[1:4], rs[3]))
	assert.Equal(t, -1, indexRequest(rs[1:4], rs[4]))
}
