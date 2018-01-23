package reliable

import (
	"github.com/qiniu/errors"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/osl"
	"io"
)

// --------------------------------------------------------------------
// type BigArray

type BigArray struct {
	data       [][]byte
	tbl        *Table
	span, rows int
}

func OpenBigArray(tbl *Table, span int) (p *BigArray, err error) {

	rows, err := tbl.Rows()
	if err != nil {
		tbl.Close()
		err = errors.Info(err, "OpenBigArray tbl.Rows failed").Detail(err)
		return
	}

	spans := int(rows)/span + 1
	data := make([][]byte, spans)
	return &BigArray{data: data, tbl: tbl, span: span, rows: int(rows)}, nil
}

func OpenBigArrfile(fnames []string, elemlen, span, allowfails int) (p *BigArray, err error) {

	tbl, err := OpenTblfile(fnames, elemlen, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenBigArrfile failed").Detail(err)
		return
	}

	return OpenBigArray(tbl, span)
}

func (p *BigArray) Underlayer() []osl.File {

	return p.tbl.Underlayer()
}

func (p *BigArray) Close() (err error) {

	return p.tbl.Close()
}

func (p *BigArray) Len() int {

	return p.rows
}

func (p *BigArray) ElemLen() int {

	return p.tbl.rowlen
}

func (p *BigArray) Shrink(n int) (err error) {

	if n >= p.rows {
		return nil
	}

	err = p.tbl.Shrink(int64(n))
	if err != nil {
		err = errors.Info(err, "BigArray.Shrink failed").Detail(err)
		return
	}

	rowlen := p.tbl.rowlen
	span := p.span
	spans := (n + span - 1) / span
	p.data = p.data[:spans]
	p.rows = n
	if spans > 0 {
		pdata := p.data[spans-1]
		if pdata != nil {
			last := (n % span) * rowlen
			if last > 0 {
				// Clear the excessive part.
				for i := last; i < len(pdata); i++ {
					pdata[i] = 0
				}
			}
		}
	}
	return
}

func (p *BigArray) Puts(ielem int, elems []byte) (err error) {

	rowlen := p.tbl.rowlen
	if len(elems)%rowlen != 0 {
		err = errors.Info(ErrInvalidArgs, "reliable.BigArray.Puts failed: invalid arguments")
		return
	}

	for i := 0; i < len(elems); i += rowlen {
		err = p.Put(ielem, elems[i:i+rowlen])
		if err != nil {
			return
		}
		ielem++
	}
	return
}

func (p *BigArray) Put(ielem int, elem []byte) (err error) {

	rowlen := p.tbl.rowlen
	spanlen := rowlen * p.span
	ispan := ielem / p.span
	ioff := (ielem % p.span) * rowlen

	err = p.tbl.WriteRow(int64(ielem), elem)
	if err != nil {
		err = errors.Info(err, "reliable.BigArray.Put failed").Detail(err)
		return
	}

	if ispan == len(p.data) {
		p.data = append(p.data, nil)
	} else if ispan > len(p.data) {
		n := ispan - len(p.data) + 1
		zeros := make([][]byte, n)
		p.data = append(p.data, zeros...)
	}

	spandata := p.data[ispan]
	if spandata == nil {
		if p.rows > ispan*p.span {
			spandata, err = p.requireSpan(ispan)
			if err != nil {
				err = errors.Info(err, "reliable.BigArray.Put failed").Detail(err)
				return
			}
		} else {
			spandata = make([]byte, spanlen)
		}
		p.data[ispan] = spandata
	}
	copy(spandata[ioff:], elem)

	if p.rows <= ielem {
		p.rows = ielem + 1
	}
	return
}

func (p *BigArray) Get(ielem int, elem []byte) (err error) {

	rowlen := p.tbl.rowlen
	if rowlen != len(elem) {
		err = errors.Info(ErrInvalidArgs, "reliable.BigArray.Get failed: invalid arguments")
		return
	}

	ispan := ielem / p.span
	ioff := (ielem % p.span) * rowlen

	if ielem >= p.rows {
		return io.EOF
	}

	spandata := p.data[ispan]
	if spandata == nil {
		spandata, err = p.requireSpan(ispan)
		if err != nil {
			err = errors.Info(err, "reliable.BigArray.Get failed").Detail(err)
			return
		}
	}
	copy(elem, spandata[ioff:])
	return
}

func (p *BigArray) Gets(ielem int, elems []byte) (n int, err error) {

	rowlen := p.tbl.rowlen
	if len(elems)%rowlen != 0 {
		err = errors.Info(ErrInvalidArgs, "reliable.BigArray.Gets failed: invalid arguments")
		return
	}

	for i := 0; i < len(elems); i += rowlen {
		err = p.Get(ielem, elems[i:i+rowlen])
		if err != nil {
			return
		}
		ielem++
		n++
	}
	return
}

func (p *BigArray) requireSpan(ispan int) (spandata []byte, err error) {

	rowlen := p.tbl.rowlen
	span := p.span

	spandata = p.data[ispan]
	if spandata != nil {
		return
	}

	spanlen := rowlen * span
	spandata = make([]byte, spanlen)

	irow := ispan * span
	ilastspan := (p.rows - 1) / span
	if ispan == ilastspan {
		nrow := p.rows - irow
		spandata = spandata[:rowlen*nrow]
	}
	err = p.tbl.ReadRows(int64(irow), spandata)
	if err != nil {
		err = errors.Info(err, "reliable.BigArray.Get loadSpan failed").Detail(err)
		return
	}
	p.data[ispan] = spandata[:spanlen]
	return
}

// --------------------------------------------------------------------
