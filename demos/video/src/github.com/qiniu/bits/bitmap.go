package bits

import (
	"syscall"
)

// -----------------------------------------------------------
// type Bitmap

type Bitmap struct {
	Data  []uint64
	ifree int
}

func NewBitmap(data []uint64) (p *Bitmap) {
	return &Bitmap{data, 0}
}

// -----------------------------------------------------------

func (p *Bitmap) Find(doClear bool) (idx int, err error) {

	for i := p.ifree; i < len(p.Data); i++ {
		v := p.Data[i]
		if v != 0 {
			p.ifree = i
			v = v &^ (v - 1)
			if doClear {
				p.Data[i] ^= v
			}
			return (i << 6) + log2(v), nil
		}
	}
	return -1, syscall.ENOENT
}

func (p *Bitmap) FindFrom(from int, doClear bool) (idx int, err error) {

	ifrom := from >> 6
	if ifrom < p.ifree {
		ifrom = p.ifree
	}

	for i := ifrom; i < len(p.Data); i++ {
		v := p.Data[i]
		if v != 0 {
			if i == ifrom {
				v &^= (1 << uint(from&0x3f)) - 1
				if v == 0 {
					continue
				}
			}
			v &^= (v - 1)
			if doClear {
				p.Data[i] ^= v
			}
			return (i << 6) + log2(v), nil
		}
	}
	return -1, syscall.ENOENT
}

func (p *Bitmap) Has(idx int) bool {

	i := idx >> 6
	if i >= len(p.Data) {
		return false
	}

	mask := uint64(1) << uint(idx&0x3f)
	return (p.Data[i] & mask) != 0
}

// -----------------------------------------------------------

func (p *Bitmap) Clear(idx int) error {

	i := idx >> 6
	if i >= len(p.Data) {
		return nil
	}

	mask := uint64(1) << uint(idx&0x3f)
	p.Data[i] &^= mask

	return nil
}

func (p *Bitmap) ClearRange(from, to int) error {

	if from > to {
		return syscall.EINVAL
	}

	ifrom := from >> 6
	ito := to >> 6
	if ito >= len(p.Data) {
		zero := make([]uint64, ito+1-len(p.Data))
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
		p.Data[ifrom] &= mfrom
		ifrom++
		for ifrom < ito {
			p.Data[ifrom] = 0
			ifrom++
		}
		p.Data[ito] &^= mto
	}

	return nil
}

// -----------------------------------------------------------

func (p *Bitmap) Set(idx int) error {

	i := idx >> 6
	if i >= len(p.Data) {
		zero := make([]uint64, i+1-len(p.Data))
		p.Data = append(p.Data, zero...)
	}

	if i < p.ifree {
		p.ifree = i
	}

	mask := uint64(1) << uint(idx&0x3f)
	p.Data[i] |= mask

	return nil
}

func (p *Bitmap) SetRange(from, to int) error {

	if from > to {
		return syscall.EINVAL
	}

	ifrom := from >> 6
	ito := to >> 6
	if ito >= len(p.Data) {
		zero := make([]uint64, ito+1-len(p.Data))
		p.Data = append(p.Data, zero...)
	}

	if ifrom < p.ifree {
		p.ifree = ifrom
	}

	mfrom := (uint64(1) << uint(from&0x3f)) - 1
	mto := (uint64(1) << uint((to+1)&0x3f)) - 1
	if mto == 0 {
		mto = ^uint64(0)
	}
	if ifrom == ito {
		p.Data[ito] |= (mfrom ^ mto)
	} else {
		p.Data[ifrom] |= ^mfrom
		ifrom++
		for ifrom < ito {
			p.Data[ifrom] = 0xffffffffffffffff
			ifrom++
		}
		p.Data[ito] |= mto
	}

	return nil
}

func (p *Bitmap) DataOf() []uint64 {
	return p.Data
}

// -----------------------------------------------------------
