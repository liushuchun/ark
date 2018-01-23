package reliable

import (
	gbytes "bytes"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/reliable/osl"
	"github.com/qiniu/ts"
	"testing"
)

// ---------------------------------------------------

func openTable(modess [][]int, rowlen, allowfails int) (p *Table, err error) {

	files := make([]osl.File, len(modess))
	for i, modes := range modess {
		files[i] = newBuffer(modes)
	}
	return OpenTable(files, rowlen, allowfails)
}

type testTableCase struct {
	data []string
	row  int64
	rows int64
}

func TestTable(t *testing.T) {

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
	p, err := openTable(modess, rowlen, 1)
	if p == nil || err != nil {
		ts.Fatal(t, "openTable failed:", err)
	}
	defer p.Close()

	rbuf := make([]byte, rowlen)
	for i, c := range cases {
		if len(c.data) == 1 {
			err = p.WriteRow(c.row, makeSlice(c.data[0], rowlen))
		} else {
			err = p.WriteRows(c.row, makeSlices(c.data, rowlen))
		}
		if err != nil {
			ts.Fatal(t, "WriteRow failed:", err, i)
		}
		rows, err := p.Rows()
		if err != nil || rows != c.rows {
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
			err = p.ReadRow(int64(row), rbuf)
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
		err = p.ReadRows(0, rbufs)
		if err != nil {
			ts.Fatal(t, "ReadRow failed:", err)
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

// ---------------------------------------------------
