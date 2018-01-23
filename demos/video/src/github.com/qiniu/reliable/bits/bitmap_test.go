package bits

import (
	bits "github.com/qiniu/bits/testing"
	"github.com/qiniu/reliable"
	"github.com/qiniu/reliable/osl"
	. "github.com/qiniu/reliable/ts"
	"github.com/qiniu/ts"
	"testing"
)

// -----------------------------------------------------------

func TestBitmap(t *testing.T) {

	modess := [][]int{
		{WriteOk, WriteBad, WriteBad, WriteOk, WriteBad, WriteBad},
		{WriteOk, WriteOk, WriteBad, WriteBad, WriteBad, WriteOk},
		{WriteOk, WriteOk, WriteOk, WriteOk, WriteOk, WriteOk},
	}

	files := make([]osl.File, len(modess))
	for i, modes := range modess {
		files[i] = NewBuffer(modes)
	}

	tbl, err := reliable.OpenTable(files, 8, 1)
	if err != nil {
		ts.Fatal(t, "OpenTable failed:", err)
	}

	b, err := OpenBitmap(tbl)
	if err != nil {
		ts.Fatal(t, "OpenBitmap failed:", err)
	}

	bits.BitmapTest(b, t)

	b.SetRange(10073, 10129)
	b.Set(20024)
	high, _, _ := bits.Count(b, 10000, 55000)

	// reopen reliable-files
	{
		tbl, err := reliable.OpenTable(files, 8, 1)
		if err != nil {
			ts.Fatal(t, "OpenTable failed:", tbl, err)
		}

		b, err := OpenBitmap(tbl)
		if err != nil {
			ts.Fatal(t, "OpenBitmap failed:", b, err)
		}

		for i := 10073; i <= 10129; i++ {
			if !b.Has(i) {
				ts.Fatal(t, "OpenBitmap, Has failed:", i)
			}
		}
		if b.Has(10072) || b.Has(10130) {
			ts.Fatal(t, "OpenBitmap, Has failed")
		}
		high2, _, _ := bits.Count(b, 10000, 55000)
		if high != high2 || high != 10129-10073+1+1 {
			ts.Fatal(t, "OpenBitmap, Count failed", high, high2)
		}
	}
}

// -----------------------------------------------------------
