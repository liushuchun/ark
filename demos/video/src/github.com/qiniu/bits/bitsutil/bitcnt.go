package bitsutil

import (
	"syscall"
)

// -----------------------------------------------------------
// func Count

func Count(data []uint64, from, to int) (high, low int, err error) {

	if from > to {
		err = syscall.EINVAL
		return
	}

	ifrom := from >> 6
	ito := to >> 6

	mfrom := (uint64(1) << uint(from&0x3f)) - 1
	mto := (uint64(1) << uint((to+1)&0x3f)) - 1
	if mto == 0 {
		mto = ^uint64(0)
	}
	if ito >= len(data) {
		ito = len(data) - 1
		mto = ^uint64(0)
	}
	for i := ifrom; i <= ito; i++ {
		v := data[i]
		if i == ifrom {
			v &= ^mfrom
		}
		if i == ito {
			v &= mto
		}
		high += bitcnt(v)
	}
	low = to - from + 1 - high
	return
}

// http://webpages.charter.net/tlikens/tech/bitmaps/bit_popcnt.html#divide_and_conquer
func bitcnt(x uint64) (count int) {
	x = x - ((x >> 1) & 0x5555555555555555)
	x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
	x = (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
	x = (x * 0x0101010101010101) >> 56
	return int(x)
}

// -----------------------------------------------------------
