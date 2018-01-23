package log

import (
	"fmt"
	"github.com/qiniu/reliable/osl"
	rts "github.com/qiniu/reliable/ts"
	"github.com/qiniu/ts"
	"io"
	"testing"
)

// ---------------------------------------------------

const (
	writeOk    = rts.WriteOk
	writeFail  = rts.WriteFail
	writeBad   = rts.WriteBad
	writeShort = rts.WriteShort
)

func newBuffer(modes []int) *rts.Buffer {

	return rts.NewBuffer(modes)
}

// --------------------------------------------------------------------

func openLogger(modess [][]int, linemax, allowfails int) (p *Logger, err error) {

	files := make([]osl.File, len(modess))
	for i, modes := range modess {
		files[i] = newBuffer(modes)
	}
	return OpenEx(files, linemax, allowfails)
}

type testLogCase struct {
	a   string
	b   int
	c   bool
	opt string
}

func TestLogger(t *testing.T) {

	modess := [][]int{
		{writeOk, writeBad, writeBad, writeOk, writeBad, writeBad, writeShort},
		{writeOk, writeOk, writeBad, writeBad, writeBad, writeOk, writeFail},
		{writeOk, writeOk, writeOk, writeOk, writeOk, writeOk, writeFail},
	}

	cases := []testLogCase{
		{"ab \tc\r\n", 123, true, ""},
		{"e fg", -1, false, "543"},
		{"cool\r\n", 5356, true, ""},
		{"high", -123, false, ""},
		{"123abc", 1235, true, "abc123"},
		{"56 abc", 4123, false, "a"},
		{" \t56 abc", 34123, false, "4a"},
	}

	p, err := openLogger(modess, 0, 1)
	if err != nil {
		ts.Fatal(t, "openLogger failed:", err)
	}
	defer p.Close()

	for i, c := range cases {
		if c.opt == "" {
			err = p.Println(c.a, c.b, c.c)
			if err != nil {
				ts.Fatal(t, "log.Println failed:", err)
			}
		} else {
			err = p.Println(c.a, c.b, c.c, c.opt)
			if err != nil {
				if i == 6 {
					break
				}
				ts.Fatal(t, "log.Println failed:", err)
			}
		}

		var va string
		var vb int
		var vc bool
		var vopt string

		lr := p.Reader(0)
		for j := 0; j <= i; j++ {
			c := cases[j]
			if c.opt == "" {
				err = lr.Scanln(&va, &vb, &vc)
				if err != nil || va != c.a || vb != c.b || vc != c.c {
					ts.Fatal(t, "log.Scanln failed:", va, vb, vc, err)
				}
			} else {
				err = lr.Scanln(&va, &vb, &vc, &vopt)
				if err != nil || va != c.a || vb != c.b || vc != c.c || vopt != c.opt {
					ts.Fatal(t, "log.Scanln failed:", va, vb, vc, vopt, err)
				}
			}
		}
		err = lr.Scanln(&va, &vb, &vc)
		if err != io.EOF {
			ts.Fatal(t, "log.Scanln failed:", i, err)
		}
	}

	{
		p2, err := OpenEx(p.files, 0, 1)
		if err != nil {
			ts.Fatal(t, "openLogger failed:", err)
		}
		defer p2.Close()

		var va string
		var vb int
		var vc bool
		var vopt string

		lr := p2.Reader(0)
		for j := 0; j < len(cases)-1; j++ {
			c := cases[j]
			if c.opt == "" {
				err = lr.Scanln(&va, &vb, &vc)
				if err != nil || va != c.a || vb != c.b || vc != c.c {
					ts.Fatal(t, "log.Scanln failed:", va, vb, vc, err)
				}
			} else {
				err = lr.Scanln(&va, &vb, &vc, &vopt)
				if err != nil || va != c.a || vb != c.b || vc != c.c || vopt != c.opt {
					ts.Fatal(t, "log.Scanln failed:", va, vb, vc, vopt, err)
				}
			}
		}
		err = lr.Scanln(&va, &vb, &vc)
		if err != io.EOF {
			ts.Fatal(t, "log.Scanln failed:", err)
		}
	}

	{
		p2, err := OpenEx(p.files, 0, 1)
		if err != nil {
			ts.Fatal(t, "openLogger failed:", err)
		}
		defer p2.Close()

		fmt.Println("==== ReadFrom begin ===========================")

		buf := make([]byte, 1024)
		n, err := p2.ReadFrom(buf, 0)
		if err != nil {
			ts.Fatal(t, "ReadFrom failed:", n, err)
		}
		text := string(buf[:n])
		text2 := `2eppkc	ab\w\tc\r\n 123 true
1a2zdbp	e\wfg -1 false 543
15qjidt	cool\r\n 5356 true
15698tk	high -123 false
1w481wk	123abc 1235 true abc123
6i5ebm	56\wabc 4123 false a
`
		fmt.Println(text)
		if text != text2 {
			ts.Fatal(t, "ReadFrom check data failed:", len(text), len(text2))
		}

		n, err = p2.ReadFrom(buf, int64(n))
		if err != io.EOF || n != 0 {
			ts.Fatal(t, "ReadFrom check eof failed:", n, err)
		}
	}
}

// --------------------------------------------------------------------
