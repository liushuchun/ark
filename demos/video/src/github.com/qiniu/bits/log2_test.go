package bits

import (
	"testing"
)

// -----------------------------------------------------------

func Test_log2(t *testing.T) {

	for i := uint(0); i < 64; i++ {
		var v = uint64(1) << i
		if log2(v) != int(i) {
			t.Fatal("log2 failed:", v, log2(v))
		}
	}
}

func TestLog2(t *testing.T) {

	for v := uint64(1); v < (1 << 9); v++ {
		i := Log2(v)
		low := uint64(1) << uint(i)
		high := uint64(1) << uint(i+1)
		if v < low || v >= high {
			t.Fatal("Log2 failed:", v, i)
		}
	}
}

// -----------------------------------------------------------

type findCase struct {
	v uint64
	i int
}

func TestFind(t *testing.T) {

	cases := []findCase{
		{0, -1},
		{1, 0},
		{2, 1},
		{3, 0},
		{4, 2},
		{5, 0},
		{6, 1},
		{7, 0},
		{8, 3},
	}
	for _, c := range cases {
		if Find(c.v) != c.i {
			t.Fatal("Find", c.v, Find(c.v), c.i)
		}
	}
}

// -----------------------------------------------------------
