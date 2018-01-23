package ta

import (
	"github.com/qiniu/encoding/binary"
	"github.com/qiniu/errors"
	"github.com/qiniu/reliable/proto"
	"syscall"
	"unsafe"
)

// --------------------------------------------------------------------
// type IBitmap

type IBitmap interface {
	proto.Bitmap
	IUnderlayer
}

// --------------------------------------------------------------------
// type Bitmap

type Bitmap struct {
	IBitmap
	ta *Transaction
	id int
}

func OpenBitmap(bitmap IBitmap, ta *Transaction, id int) (p *Bitmap, err error) {

	p = &Bitmap{bitmap, ta, id}

	err = ta.init(id, p)
	if err != nil {
		err = errors.Info(err, "ta.OpenBitmap failed", id).Detail(err)
	}
	return
}

func (p *Bitmap) Clear(idx int) (err error) {

	err = p.saveBits(idx, idx)
	if err != nil {
		err = errors.Info(err, "ta.Bitmap.Clear failed", idx).Detail(err)
		return
	}

	return p.IBitmap.Clear(idx)
}

func (p *Bitmap) ClearRange(from, to int) (err error) {

	err = p.saveBits(from, to)
	if err != nil {
		err = errors.Info(err, "ta.Bitmap.ClearRange failed", from, to).Detail(err)
		return
	}

	return p.IBitmap.ClearRange(from, to)
}

func (p *Bitmap) Set(idx int) (err error) {

	err = p.saveBits(idx, idx)
	if err != nil {
		err = errors.Info(err, "ta.Bitmap.Set failed", idx).Detail(err)
		return
	}

	return p.IBitmap.Set(idx)
}

func (p *Bitmap) SetRange(from, to int) (err error) {

	err = p.saveBits(from, to)
	if err != nil {
		err = errors.Info(err, "ta.Bitmap.SetRange failed", from, to).Detail(err)
		return
	}

	return p.IBitmap.SetRange(from, to)
}

func (p *Bitmap) Find(doClear bool) (idx int, err error) {

	idx, err = p.IBitmap.Find(false)
	if err != nil || !doClear {
		return
	}

	err = p.Clear(idx)
	return
}

func (p *Bitmap) FindFrom(from int, doClear bool) (idx int, err error) {

	idx, err = p.IBitmap.FindFrom(from, false)
	if err != nil || !doClear {
		return
	}

	err = p.Clear(idx)
	return
}

func (p *Bitmap) saveBits(from, to int) (err error) {

	if from > to {
		return syscall.EINVAL
	}

	data := p.DataOf()
	ifrom := from >> 6
	ito := to >> 6

	var saved []uint64
	if ifrom >= len(data) {
		saved = make([]uint64, ito+1-ifrom)
	} else if ito >= len(data) {
		saved = make([]uint64, ito+1-ifrom)
		copy(saved, data[ifrom:])
	} else {
		saved = data[ifrom : ito+1]
	}

	rl, hint, err := p.ta.beginRlog(p.id)
	if err != nil {
		err = errors.Info(err, "ta.Bitmap.saveBits: beginRlog failed").Detail(err)
		return
	}
	rl.PutUint32(uint32(ifrom))
	rl.Write(bytesOf(saved))
	rl.end(hint)
	return
}

func (p *Bitmap) DoAct(act []byte) (err error) {

	bitmap := p.IBitmap

	ifrom := int(binary.LittleEndian.Uint32(act))
	from := ifrom << 6
	data := uint64sOf(act[4:])
	for k, v := range data {
		for i := 0; i < 64; i++ {
			idx := from + k*64 + i
			if v>>uint(i)&1 > 0 {
				bitmap.Set(idx)
			} else {
				bitmap.Clear(idx)
			}
		}
	}

	return
}

// --------------------------------------------------------------------

func bytesOf(v []uint64) []byte {

	b := (*[1 << 30]byte)(unsafe.Pointer(&v[0]))
	return b[:len(v)<<3]
}

func uint64sOf(v []byte) []uint64 {

	b := (*[1 << 27]uint64)(unsafe.Pointer(&v[0]))
	return b[:len(v)>>3]
}

// --------------------------------------------------------------------
