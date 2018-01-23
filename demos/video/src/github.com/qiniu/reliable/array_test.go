package reliable

import (
	gbytes "bytes"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"testing"
)

// --------------------------------------------------------------------

func TestArray(t *testing.T) {

	modess := [][]int{
		{writeOk, writeBad, writeBad, writeOk, writeBad, writeBad},
		{writeOk, writeOk, writeBad, writeBad, writeBad, writeOk},
		{writeOk, writeOk, writeOk, writeOk, writeOk, writeOk},
	}

	maxrow := 0
	realworld := make([]string, 10)
	cases := []testTableCase{
		{[]string{"Hello"}, 3, 4},
		{[]string{"World!"}, 0, 4},
		{[]string{"xsw"}, 1, 4},
		{[]string{"abcd"}, 3, 4},
		{[]string{"ef"}, 2, 4},
		{[]string{"hi", "qiniu", "!"}, 3, 6},
	}

	rowlen := 8
	tbl, err := openTable(modess, rowlen, 1)
	if tbl == nil || err != nil {
		ts.Fatal(t, "openTable failed:", err)
	}
	p, err := OpenArray(tbl)
	if err != nil {
		ts.Fatal(t, "openArray failed:", err)
	}
	defer p.Close()

	rbuf := make([]byte, rowlen)
	for i, c := range cases {
		if len(c.data) == 1 {
			err = p.Put(int(c.row), makeSlice(c.data[0], rowlen))
		} else {
			err = p.Puts(int(c.row), makeSlices(c.data, rowlen))
		}
		if err != nil {
			ts.Fatal(t, "WriteRow failed:", err, i)
		}
		rows := p.Len()
		if rows != int(c.rows) {
			ts.Fatal(t, "Rows:", rows, err, i)
		}
		irow := int(c.row)
		for row, data := range c.data {
			realworld[irow+row] = data
		}
		if maxrow < irow+len(c.data) {
			maxrow = irow + len(c.data)
		}
		log.Println("rows:", realworld)
		for row := 0; row < maxrow; row++ {
			rdata := realworld[row]
			err = p.Get(row, rbuf)
			if err != nil {
				ts.Fatal(t, "ReadRow failed:", row, err)
			}
			rexp := makeSlice(rdata, rowlen)
			if !gbytes.Equal(rbuf, rexp) {
				ts.Fatal(t, "ReadRow failed: unexpected -", row, rbuf, rexp)
			}
		}

		log.Println("== ReadRows begin ==")
		rbufs := make([]byte, rowlen*maxrow)
		nrow, err := p.Gets(0, rbufs)
		if err != nil || nrow != maxrow {
			ts.Fatal(t, "ReadRow failed:", nrow, err)
		}
		log.Println("bufs:", rbufs)
		for row := 0; row < maxrow; row++ {
			rdata := realworld[row]
			rexp := makeSlice(rdata, rowlen)
			off := row * rowlen
			if !gbytes.Equal(rbufs[off:off+rowlen], rexp) {
				ts.Fatal(t, "ReadRows failed: unexpected -", row, rbuf, rexp)
			}
		}
	}
}

// --------------------------------------------------------------------
