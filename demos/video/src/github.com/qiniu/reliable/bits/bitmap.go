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
// type Bitmap

type Bitmap struct {
	bitmap *bits.Bitmap
	tbl    *reliable.Table
}

func OpenBitmap(tbl *reliable.Table) (p *Bitmap, err error) {

	rowlen := tbl.RowLen()
	if (rowlen & 7) != 0 {
		return nil, syscall.EINVAL
	}

	rows, err := tbl.Rows()
	if err != nil {
		tbl.Close()
		err = errors.Info(err, "OpenBitmap tbl.Rows failed").Detail(err)
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
			err = errors.Info(err, "OpenBitmap tbl.ReadRows failed").Detail(err)
			return
		}
	}
	return &Bitmap{bitmap: bits.NewBitmap(bitmap), tbl: tbl}, nil
}

func OpenBitfile(fnames []string, elemlen, allowfails int) (p *Bitmap, err error) {

	tbl, err := reliable.OpenTblfile(fnames, elemlen, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenBitfile failed").Detail(err)
		return
	}

	return OpenBitmap(tbl)
}

func (p *Bitmap) Close() (err error) {

	return p.tbl.Close()
}

func (p *Bitmap) Underlayer() []osl.File {

	return p.tbl.Underlayer()
}

// --------------------------------------------------------------------

func (p *Bitmap) update(idx int) error {

	if (idx >> 6) >= len(p.bitmap.Data) {
		return nil
	}

	rowlen := p.tbl.RowLen()
	row := (idx >> 3) / rowlen
	start := row * rowlen
	buf := (*[1 << 30]byte)(unsafe.Pointer(&p.bitmap.Data[0]))

	return p.tbl.WriteRow(int64(row), buf[start:start+rowlen])
}

func (p *Bitmap) updateRange(from, to int) error {

	rowlen := p.tbl.RowLen()
	rowfrom := (from >> 3) / rowlen
	rowto := (to >> 3) / rowlen

	start := rowfrom * rowlen
	end := (rowto + 1) * rowlen
	buf := (*[1 << 30]byte)(unsafe.Pointer(&p.bitmap.Data[0]))

	return p.tbl.WriteRows(int64(rowfrom), buf[start:end])
}

// --------------------------------------------------------------------

func (p *Bitmap) Find(doClear bool) (idx int, err error) {

	idx, err = p.bitmap.Find(doClear)
	if doClear && err == nil {
		err = p.update(idx)
	}
	return
}

func (p *Bitmap) FindFrom(from int, doClear bool) (idx int, err error) {

	idx, err = p.bitmap.FindFrom(from, doClear)
	if doClear && err == nil {
		err = p.update(idx)
	}
	return
}

func (p *Bitmap) Has(idx int) bool {

	return p.bitmap.Has(idx)
}

func (p *Bitmap) Clear(idx int) error {

	err := p.bitmap.Clear(idx)
	if err == nil {
		err = p.update(idx)
	}
	return err
}

func (p *Bitmap) ClearRange(from, to int) error {

	err := p.bitmap.ClearRange(from, to)
	if err == nil {
		err = p.updateRange(from, to)
	}
	return err
}

func (p *Bitmap) Set(idx int) error {

	err := p.bitmap.Set(idx)
	if err == nil {
		err = p.update(idx)
	}
	return err
}

func (p *Bitmap) SetRange(from, to int) error {

	err := p.bitmap.SetRange(from, to)
	if err == nil {
		err = p.updateRange(from, to)
	}
	return err
}

func (p *Bitmap) DataOf() []uint64 {
	return p.bitmap.Data
}

// --------------------------------------------------------------------
