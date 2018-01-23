package ns

import (
	"github.com/qiniu/reliable/osl"
	. "github.com/qiniu/reliable/ts"
	"github.com/qiniu/ts"
	"testing"
)

// --------------------------------------------------------------------

func TestNS(t *testing.T) {

	modess := [][]int{
		{WriteOk, WriteBad, WriteBad, WriteOk, WriteBad, WriteBad},
		{WriteOk, WriteOk, WriteBad, WriteBad, WriteBad, WriteOk},
		{WriteOk, WriteOk, WriteOk, WriteOk, WriteOk, WriteOk},
	}

	files := make([]osl.File, len(modess))
	for i, modes := range modess {
		files[i] = NewBuffer(modes)
	}

	p, err := Open(files, 0, 1)
	if err != nil {
		ts.Fatal(t, "openNs failed:", err)
	}
	defer p.Close()

	aaa, err := p.Register("aaa")
	if err != nil {
		ts.Fatal(t, "Register failed:", err)
	}

	bbb, err := p.Register("bbbbb")
	if err != nil {
		ts.Fatal(t, "Register failed:", err)
	}

	aaa2, err := p.Register("aaa")
	if err != nil || aaa != aaa2 {
		ts.Fatal(t, "Register failed:", aaa2, err)
	}

	bbb2, err := p.Register("bbbbb")
	if err != nil || bbb != bbb2 {
		ts.Fatal(t, "Register failed:", bbb2, err)
	}

	p.Close()

	p, err = Open(files, 0, 1)
	if err != nil {
		ts.Fatal(t, "openNs failed:", err)
	}

	aaa3, err := p.Register("aaa")
	if err != nil || aaa != aaa3 {
		ts.Fatal(t, "Register failed:", aaa3, err)
	}

	bbb3, err := p.Register("bbbbb")
	if err != nil || bbb != bbb3 {
		ts.Fatal(t, "Register failed:", bbb3, err)
	}
}

// --------------------------------------------------------------------
