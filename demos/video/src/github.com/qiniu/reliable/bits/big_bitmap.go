package bits

import (
	"github.com/qiniu/bits"
	"github.com/qiniu/errors"
	"github.com/qiniu/reliable"
	"github.com/qiniu/reliable/osl"
	"syscall"
	"unsafe"
)

// --------------------------------------------------------------------
// type BigBitmap

type BigBitmap struct {
	bitmap *bits.BigBitmap
	tbl    *reliable.Table
}

func OpenBigBitmap(tbl *reliable.Table, maxbits int) (p *BigBitmap, err error) {

	rowlen := tbl.RowLen()
	if (rowlen & 7) != 0 {
		return nil, syscall.EINVAL
	}

	rows, err := tbl.Rows()
	if err != nil {
		tbl.Close()
		err = errors.Info(err, "OpenBigBitmap tbl.Rows failed").Detail(err)
		return
	}

	var bitmap []uint64
	if rows > 0 {
		datalen := int(rows) * rowlen
		bitmap = make([]uint64, datalen>>3)
		pp := (*[1 << 30]byte)(unsafe.Pointer(&bitmap[0]))
		err = tbl.ReadRows(0, pp[:datalen])
		if err != nil {
			tbl.Close()
			err = errors.Info(err, "OpenBigBitmap tbl.ReadRows failed").Detail(err)
			return
		}
	}

	b := bits.NewBigBitmap(bitmap, maxbits)
	return &BigBitmap{bitmap: b, tbl: tbl}, nil
}

func OpenBigBitfile(fnames []string, elemlen, maxbits, allowfails int) (p *BigBitmap, err error) {

	tbl, err := reliable.OpenTblfile(fnames, elemlen, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenBitfile failed").Detail(err)
		return
	}

	return OpenBigBitmap(tbl, maxbits)
}

func (p *BigBitmap) Close() (err error) {

	return p.tbl.Close()
}

func (p *BigBitmap) Underlayer() []osl.File {

	return p.tbl.Underlayer()
}

// --------------------------------------------------------------------

func (p *BigBitmap) update(idx int) error {

	if (idx >> 6) >= len(p.bitmap.Data) {
		return nil
	}

	rowlen := p.tbl.RowLen()
	row := (idx >> 3) / rowlen
	start := row * rowlen
	buf := (*[1 << 30]byte)(unsafe.Pointer(&p.bitmap.Data[0]))

	return p.tbl.WriteRow(int64(row), buf[start:start+rowlen])
}

func (p *BigBitmap) updateRange(from, to int) error {

	rowlen := p.tbl.RowLen()
	rowfrom := (from >> 3) / rowlen
	rowto := (to >> 3) / rowlen

	start := rowfrom * rowlen
	end := (rowto + 1) * rowlen
	buf := (*[1 << 30]byte)(unsafe.Pointer(&p.bitmap.Data[0]))

	return p.tbl.WriteRows(int64(rowfrom), buf[start:end])
}

// --------------------------------------------------------------------

func (p *BigBitmap) Find(doClear bool) (idx int, err error) {

	idx, err = p.bitmap.Find(doClear)
	if doClear && err == nil {
		err = p.update(idx)
	}
	return
}

func (p *BigBitmap) FindFrom(from int, doClear bool) (idx int, err error) {

	idx, err = p.bitmap.FindFrom(from, doClear)
	if doClear && err == nil {
		err = p.update(idx)
	}
	return
}

func (p *BigBitmap) Has(idx int) bool {

	return p.bitmap.Has(idx)
}

func (p *BigBitmap) Clear(idx int) error {

	err := p.bitmap.Clear(idx)
	if err == nil {
		err = p.update(idx)
	}
	return err
}

func (p *BigBitmap) ClearRange(from, to int) error {

	err := p.bitmap.ClearRange(from, to)
	if err == nil {
		err = p.updateRange(from, to)
	}
	return err
}

func (p *BigBitmap) Set(idx int) error {

	err := p.bitmap.Set(idx)
	if err == nil {
		err = p.update(idx)
	}
	return err
}

func (p *BigBitmap) SetRange(from, to int) error {

	err := p.bitmap.SetRange(from, to)
	if err == nil {
		err = p.updateRange(from, to)
	}
	return err
}

func (p *BigBitmap) DataOf() []uint64 {
	return p.bitmap.Data
}

// --------------------------------------------------------------------
