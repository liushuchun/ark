package bits

// --------------------------------------------------------------------

func Counts(v []uint64) (count int) {

	for _, x := range v {
		count += Count(x)
	}
	return
}

// http://webpages.charter.net/tlikens/tech/bitmaps/bit_popcnt.html#divide_and_conquer
//
func Count(x uint64) (count int) {

	x = x - ((x >> 1) & 0x5555555555555555)
	x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
	x = (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
	x = (x * 0x0101010101010101) >> 56

	return int(x)
}

// --------------------------------------------------------------------
