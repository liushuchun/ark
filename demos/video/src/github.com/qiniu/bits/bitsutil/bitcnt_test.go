package bitsutil

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestCount(t *testing.T) {
	type testcase struct {
		data     []uint64
		from, to int
		high     int
	}
	tcs := []testcase{
		{
			[]uint64{^uint64(0), ^uint64(0)},
			0, 31,
			32,
		},
		{
			[]uint64{^uint64(0), ^uint64(0)},
			32, 63,
			32,
		},
		{
			[]uint64{^uint64(0), ^uint64(0)},
			0, 63,
			64,
		},
		{
			[]uint64{^uint64(0), ^uint64(0)},
			64, 127,
			64,
		},
		{
			[]uint64{^uint64(0), ^uint64(0)},
			64, 128,
			64,
		},
	}

	for _, tc := range tcs {
		high, _, err := Count(tc.data, tc.from, tc.to)
		if err != nil {
			t.Error(err)
		}
		if high != tc.high {
			t.Error(tc, high)
		}
	}
}

func TestBitcnt(t *testing.T) {
	oldBitcnt := func(x uint64) (count int) {
		for ; x > 0; count++ {
			x &= x - 1
		}
		return
	}

	x := ^uint64(0)
	if a, b := oldBitcnt(x), bitcnt(x); a != b {
		t.Error(x, a, b)
	}

	for i := 0; i < 100; i++ {
		x := (uint64(rand.Uint32()) << 32) | uint64(rand.Uint32())
		if a, b := oldBitcnt(x), bitcnt(x); a != b {
			t.Error(x, a, b)
		}
	}
}
