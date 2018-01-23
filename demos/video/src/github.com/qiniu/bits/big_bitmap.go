package bits

import (
	"errors"
	"syscall"
)

// -----------------------------------------------------------
// type BigBitmap

var ErrOutOfRange = errors.New("out of range")

type BigBitmap struct {
	Data      []uint64
	indexs    []uint64
	min, ncap int
}

// -----------------------------------------------------------

const minCap = 1 << 12

func NewBigBitmap(data []uint64, maxbits int) (p *BigBitmap) { // 暂时不考虑动态扩展

	if maxbits < minCap {
		maxbits = minCap
	}
	level := Log2(uint64(maxbits-1)) / 6

	min := 1
	for i := 2; i <= level; i++ {
		min = (min << 6) + 1
	}
	max := (min << 6) + 1
	indexs := make([]uint64, min)

	p = &BigBitmap{Data: data, indexs: indexs, min: min, ncap: max - min}
	if len(data) > 0 {
		if len(data) > p.ncap {
			panic("NewBigBitmap: out of range")
		}
		p.updateRange(0, len(data)-1)
	}
	return p
}

func (p *BigBitmap) updateRange(ifrom, ito int) {

	min := p.min
	data := p.Data
	indexs := p.indexs

	for {
		spanfrom := ifrom >> 6
		spanto := ito >> 6
		ifrom = -1
		ito = -1
		for span := spanfrom; span <= spanto; span++ {
			if idx := updateSpan(indexs, data, min, span<<6); idx >= 0 {
				ifrom, ito = idx, idx
				for {
					span++
					if span > spanto {
						break
					}
					if idx := updateSpan(indexs, data, min, span<<6); idx >= 0 {
						ito = idx
					}
				}
				break
			}
		}
		if ifrom <= 0 {
			break
		}
		min = (min - 1) >> 6
		data = indexs[min:]
		ifrom -= min
		ito -= min
	}
}

func updateSpan(indexs, data []uint64, min, i int) int {

	var newidx uint64

	for j := 0; j < 64 && i+j < len(data); j++ {
		if data[i+j] != 0 {
			newidx |= uint64(1) << uint(j)
		}
	}
	i = (min + i - 1) >> 6
	if indexs[i] != newidx {
		indexs[i] = newidx
		return i
	}
	return -1
}

func clearFlag(indexs []uint64, i int, idx int) {

	for {
		i = (i - 1) >> 6
		idx >>= 6
		indexs[i] &= ^(uint64(1) << uint(idx&0x3f))
		if i <= 0 || indexs[i] != 0 {
			break
		}
	}
}

func setFlag(indexs []uint64, oldv uint64, i int, idx int) {

	for oldv == 0 && i > 0 {
		i = (i - 1) >> 6
		idx >>= 6
		oldv = indexs[i]
		indexs[i] |= uint64(1) << uint(idx&0x3f)
	}
}

// -----------------------------------------------------------

func (p *BigBitmap) Has(idx int) bool {

	i := idx >> 6
	if i >= len(p.Data) {
		return false
	}

	mask := uint64(1) << uint(idx&0x3f)
	return (p.Data[i] & mask) != 0
}

func (p *BigBitmap) Find(doClear bool) (idx int, err error) {

	return p.find(0, doClear)
}

func (p *BigBitmap) find(i int, doClear bool) (idx int, err error) {

	var v uint64

	for i < p.min {
		v = p.indexs[i]
		if v == 0 {
			return -1, syscall.ENOENT
		}
		i = (i << 6) + 1 + Find(v)
	}

	i -= p.min
	v = p.Data[i]

	v &^= (v - 1)
	idx = (i << 6) + log2(v)

	if doClear {
		p.Data[i] &= ^v
		if p.Data[i] == 0 {
			clearFlag(p.indexs, p.min+i, idx)
		}
	}
	return
}

func (p *BigBitmap) FindFrom(from int, doClear bool) (idx int, err error) {

	i := from >> 6
	if i >= len(p.Data) {
		return -1, syscall.ENOENT
	}

	v := p.Data[i] &^ ((1 << uint(from&0x3f)) - 1)
	if v != 0 {
		v &^= v - 1
		idx = (i << 6) + log2(v)
		if doClear {
			p.Data[i] &= ^v
			if p.Data[i] == 0 {
				clearFlag(p.indexs, p.min+i, idx)
			}
		}
		return
	}

	indexs := p.indexs
	max := p.min

	for max > 0 {
		min := (max - 1) >> 6
		from = i + 1
		i = from >> 6
		if min+i >= max {
			break
		}
		v = indexs[min+i] &^ ((1 << uint(from&0x3f)) - 1)
		if v == 0 {
			max = min
			continue
		}
		i = ((min + i) << 6) + 1 + Find(v)
		return p.find(i, doClear)
	}
	return -1, syscall.ENOENT
}

// -----------------------------------------------------------

func makeZero(n int) []uint64 {

	return make([]uint64, (n+63)&^63)
}

func (p *BigBitmap) Set(idx int) error {

	i := idx >> 6
	if i >= len(p.Data) {
		if i >= p.ncap {
			return ErrOutOfRange
		}
		zero := makeZero(i + 1 - len(p.Data))
		p.Data = append(p.Data, zero...)
	}

	mask := uint64(1) << uint(idx&0x3f)
	oldv := p.Data[i]
	p.Data[i] |= mask

	setFlag(p.indexs, oldv, p.min+i, idx)
	return nil
}

func (p *BigBitmap) Clear(idx int) error {

	i := idx >> 6
	if i >= len(p.Data) {
		return nil
	}

	mask := uint64(1) << uint(idx&0x3f)
	if (p.Data[i] & mask) != 0 {
		p.Data[i] &^= mask
		if p.Data[i] == 0 {
			clearFlag(p.indexs, p.min+i, idx)
		}
	}
	return nil
}

// -----------------------------------------------------------

func (p *BigBitmap) ClearRange(from, to int) error {

	if from > to {
		return syscall.EINVAL
	}

	ifrom := from >> 6
	ito := to >> 6
	if ito >= len(p.Data) {
		if ito >= p.ncap {
			return ErrOutOfRange
		}
		zero := makeZero(ito + 1 - len(p.Data))
		p.Data = append(p.Data, zero...)
	}

	mfrom := (uint64(1) << uint(from&0x3f)) - 1
	mto := (uint64(1) << uint((to+1)&0x3f)) - 1
	if mto == 0 {
		mto = ^uint64(0)
	}
	if ifrom == ito {
		p.Data[ito] &^= (mfrom ^ mto)
	} else {
		ifrom := ifrom
		p.Data[ifrom] &= mfrom
		ifrom++
		for ifrom < ito {
			p.Data[ifrom] = 0
			ifrom++
		}
		p.Data[ito] &^= mto
	}

	p.updateRange(ifrom, ito)
	return nil
}

// -----------------------------------------------------------

func (p *BigBitmap) SetRange(from, to int) error {

	if from > to {
		return syscall.EINVAL
	}

	ifrom := from >> 6
	ito := to >> 6
	if ito >= len(p.Data) {
		if ito >= p.ncap {
			return ErrOutOfRange
		}
		zero := makeZero(ito + 1 - len(p.Data))
		p.Data = append(p.Data, zero...)
	}

	mfrom := (uint64(1) << uint(from&0x3f)) - 1
	mto := (uint64(1) << uint((to+1)&0x3f)) - 1
	if mto == 0 {
		mto = ^uint64(0)
	}
	if ifrom == ito {
		p.Data[ito] |= (mfrom ^ mto)
	} else {
		ifrom := ifrom
		p.Data[ifrom] |= ^mfrom
		ifrom++
		for ifrom < ito {
			p.Data[ifrom] = 0xffffffffffffffff
			ifrom++
		}
		p.Data[ito] |= mto
	}

	p.updateRange(ifrom, ito)
	return nil
}

func (p *BigBitmap) DataOf() []uint64 {
	return p.Data
}

// -----------------------------------------------------------
