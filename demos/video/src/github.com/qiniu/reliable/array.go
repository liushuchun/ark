package reliable

import (
	"github.com/qiniu/errors"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/osl"
	"io"
)

// --------------------------------------------------------------------
// type Array

type Array struct {
	data []byte
	tbl  *Table
}

func OpenArray(tbl *Table) (p *Array, err error) {

	rows, err := tbl.Rows()
	if err != nil {
		tbl.Close()
		err = errors.Info(err, "OpenArray tbl.Rows failed").Detail(err)
		return
	}

	var data []byte
	if rows > 0 {
		datalen := int(rows) * tbl.rowlen
		data = make([]byte, datalen)
		err = tbl.ReadRows(0, data)
		if err != nil {
			tbl.Close()
			err = errors.Info(err, "OpenArray tbl.ReadRows failed").Detail(err)
			return
		}
	}
	return &Array{data: data, tbl: tbl}, nil
}

func OpenArrfile(fnames []string, elemlen, allowfails int) (p *Array, err error) {

	tbl, err := OpenTblfile(fnames, elemlen, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenArrfile failed").Detail(err)
		return
	}

	return OpenArray(tbl)
}

func (p *Array) Underlayer() []osl.File {

	return p.tbl.Underlayer()
}

func (p *Array) Close() (err error) {

	return p.tbl.Close()
}

func (p *Array) Len() int {

	return len(p.data) / p.tbl.rowlen
}

func (p *Array) ElemLen() int {

	return p.tbl.rowlen
}

func (p *Array) Shrink(n int) (err error) {

	n1 := len(p.data) / p.tbl.rowlen
	if n >= n1 {
		return nil
	}

	err = p.tbl.Shrink(int64(n))
	if err != nil {
		err = errors.Info(err, "Array.Shrink failed").Detail(err)
		return
	}

	p.data = p.data[:n*p.tbl.rowlen]
	return
}

func (p *Array) Puts(ielem int, elems []byte) (err error) {

	rowlen := p.tbl.rowlen
	ioff := ielem * rowlen

	err = p.tbl.WriteRows(int64(ielem), elems)
	if err != nil {
		err = errors.Info(err, "reliable.Array.Puts failed").Detail(err)
		return
	}

	iend := ioff + len(elems)
	if len(p.data) < iend {
		if len(p.data) == ioff {
			p.data = append(p.data, elems...)
			return
		}
		zero := make([]byte, iend-len(p.data))
		p.data = append(p.data, zero...)
	}

	copy(p.data[ioff:], elems)
	return
}

func (p *Array) Put(ielem int, elem []byte) (err error) {

	rowlen := p.tbl.rowlen
	ioff := ielem * rowlen

	err = p.tbl.WriteRow(int64(ielem), elem)
	if err != nil {
		err = errors.Info(err, "reliable.Array.Put failed").Detail(err)
		return
	}

	if len(p.data) <= ioff {
		if len(p.data) == ioff {
			p.data = append(p.data, elem...)
			return
		}
		zero := make([]byte, ioff+rowlen-len(p.data))
		p.data = append(p.data, zero...)
	}

	copy(p.data[ioff:], elem)
	return
}

func (p *Array) Get(ielem int, elem []byte) (err error) {

	rowlen := p.tbl.rowlen
	if rowlen != len(elem) {
		err = errors.Info(ErrInvalidArgs, "reliable.Array.Get failed: invalid arguments")
		return
	}

	ioff := ielem * rowlen
	if ioff < len(p.data) {
		copy(elem, p.data[ioff:])
	} else {
		err = io.EOF
	}
	return
}

func (p *Array) Gets(ielem int, elems []byte) (n int, err error) {

	rowlen := p.tbl.rowlen
	if len(elems)%rowlen != 0 {
		err = errors.Info(ErrInvalidArgs, "reliable.Array.Gets failed: invalid arguments")
		return
	}

	ioff := ielem * rowlen
	if ioff < len(p.data) {
		ncopy := copy(elems, p.data[ioff:])
		if ncopy != len(elems) {
			err = io.EOF
		}
		n = ncopy / rowlen
	} else {
		err = io.EOF
	}
	return
}

// --------------------------------------------------------------------
