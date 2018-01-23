package testing

import (
	"github.com/qiniu/bits/bitsutil"
	"github.com/qiniu/ts"
	"testing"
)

// --------------------------------------------------------------------
// type Bitmap

type Bitmap interface {
	Clear(idx int) error
	ClearRange(from, to int) error
	Set(idx int) error
	SetRange(from, to int) error
	Find(doClear bool) (idx int, err error)
	FindFrom(from int, doClear bool) (idx int, err error)
	DataOf() []uint64
	Has(idx int) bool
}

// -----------------------------------------------------------

func Count(b Bitmap, from, to int) (int, int, error) {
	return bitsutil.Count(b.DataOf(), from, to)
}

func BitmapTest(b Bitmap, t *testing.T) {

	b.Set(69)
	if !b.Has(69) {
		ts.Fatal(t, "Set 69 failed")
	}
	if high, low, err := Count(b, 69, 2500); err != nil || high != 1 || low != (2500-69+1-1) {
		ts.Fatal(t, "Cal 69-2500 failed:", err, high, low)
	}

	if idx, err := b.FindFrom(0, false); err != nil || idx != 69 {
		ts.Fatal(t, "FindFrom 0 69 failed", idx, err)
	}
	if idx, err := b.FindFrom(69, false); err != nil || idx != 69 {
		ts.Fatal(t, "FindFrom 69 69 failed", idx, err)
	}
	if idx, err := b.FindFrom(70, false); err == nil {
		ts.Fatal(t, "FindFrom 70 69 failed", idx, err)
	}
	if idx, err := b.FindFrom(69, true); err != nil || idx != 69 {
		ts.Fatal(t, "FindFrom 69 69 failed", idx, err)
	}
	if idx, err := b.FindFrom(69, false); err == nil {
		ts.Fatal(t, "FindFrom 69 69 failed", idx, err)
	}

	b.Set(69)

	if idx, err := b.Find(false); err != nil || idx != 69 {
		ts.Fatal(t, "Find 69 failed:", idx, err)
	}
	if idx, err := b.Find(true); err != nil || idx != 69 {
		ts.Fatal(t, "Find 69 failed:", idx, err)
	}
	if idx, err := b.Find(true); err == nil {
		ts.Fatal(t, "Find 69 failed:", idx, err)
	}

	b.Set(69)
	b.Set(62)
	high, low, err := Count(b, 62, 69)
	if err != nil || high != 2 || low != (69-62+1-high) {
		ts.Fatal(t, "cal 62-69 failed:", err, high, low)
	}

	if idx, err := b.FindFrom(62, false); err != nil || idx != 62 {
		ts.Fatal(t, "FindFrom 62 62 failed:", idx, err)
	}
	if idx, err := b.FindFrom(63, false); err != nil || idx != 69 {
		ts.Fatal(t, "FindFrom 63 69 failed:", idx, err)
	}
	if idx, err := b.FindFrom(62, true); err != nil || idx != 62 {
		ts.Fatal(t, "FindFrom 62, 62 failed:", idx, err)
	}
	if idx, err := b.FindFrom(62, true); err != nil || idx != 69 {
		ts.Fatal(t, "FindFrom 62, 69 failed:", idx, err)
	}

	b.Set(69)
	b.Set(62)
	if idx, err := b.Find(true); err != nil || idx != 62 {
		ts.Fatal(t, "Find 62 failed:", idx, err)
	}
	if idx, err := b.Find(true); err != nil || idx != 69 {
		ts.Fatal(t, "Find 69 failed:", idx, err)
	}

	b.SetRange(69, 73)
	high, low, err = Count(b, 68, 73)
	if err != nil || high != 5 || low != (73-68+1-high) {
		ts.Fatal(t, "cal 68-73 failed:", err, high, low)
	}
	for i := 69; i <= 73; i++ {
		if idx, err := b.Find(true); err != nil || idx != i {
			ts.Fatal(t, "SetRange failed:", i, idx, err)
		}
	}
	if idx, err := b.Find(true); err == nil {
		ts.Fatal(t, "SetRange failed:", idx, err)
	}

	b.SetRange(62, 69)
	for i := 62; i <= 69; i++ {
		if idx, err := b.Find(true); err != nil || idx != i {
			ts.Fatal(t, "SetRange failed:", i, idx, err)
		}
	}
	if idx, err := b.Find(true); err == nil {
		ts.Fatal(t, "SetRange failed:", idx, err)
	}

	b.SetRange(69, 73)
	b.ClearRange(69, 73)
	if idx, err := b.Find(true); err == nil {
		ts.Fatal(t, "ClearRange failed:", idx, err)
	}

	b.SetRange(62, 69)
	b.ClearRange(62, 69)
	if idx, err := b.Find(true); err == nil {
		ts.Fatal(t, "SetRange failed:", idx, err)
	}

	b.SetRange(62, 69)
	b.ClearRange(63, 68)
	if idx, err := b.Find(true); err != nil || idx != 62 {
		ts.Fatal(t, "Find 62 failed:", idx, err)
	}
	if idx, err := b.Find(true); err != nil || idx != 69 {
		ts.Fatal(t, "Find 69 failed:", idx, err)
	}

	b.SetRange(69, 73)
	b.ClearRange(70, 73)
	if idx, err := b.Find(true); err != nil || idx != 69 {
		ts.Fatal(t, "ClearRange failed:", idx, err)
	}

	b.Set(65)
	b.SetRange(1024, 1028)
	b.Set(2048)
	high, _, _ = Count(b, 1023, 2500)
	if high != 6 {
		ts.Fatal(t, "Cal 1023-2500 failed:", err, high)
	}
	high1, _, _ := Count(b, 1, 66)
	high2, _, _ := Count(b, 67, 1026)
	high3, _, _ := Count(b, 1027, 2500)
	highAll, _, _ := Count(b, 1, 2500)
	if highAll != high1+high2+high3 || highAll != 1+(1028-1024+1)+1 {
		ts.Fatal(t, "Cal sum failed:", highAll, high1, high2, high3)
	}

	// Test set range from somewhere to 64 bit's boundary.
	testSetRange(b, t, 20, 63)
	testSetRange(b, t, 20, 127)
	testSetRange(b, t, 20, 191)

	// Test clear range from somewhere to 64 bit's boundary.
	testClearRange(b, t, 12, 20, 63)
	testClearRange(b, t, 12, 20, 127)
	testClearRange(b, t, 12, 20, 191)

	testFindFromWithIndex(b, t, 64*64+63)
	testFindFromWithIndex(b, t, 64*64*64+63)
	testFindFromWithIndex(b, t, 64*64*64*64+63)
	testBigRange(b, t, 64+63)
	testBigRange(b, t, 64*64+63)
	testBigRange(b, t, 64*64*64+63)
	testBigRange(b, t, 64*64*64*64+63)
	b.ClearRange(0, 1<<25)
}

func testSetRange(b Bitmap, t *testing.T, from, to int) {

	b.ClearRange(0, 1024)

	b.SetRange(from, to)

	if high, _, _ := Count(b, 0, 1024); high != to-from+1 {
		ts.Fatalf(t, "SetRange(%v,%v) unexpected setted count %v\n", from, to, high)
	}

	for i := from; i <= to; i++ {
		if !b.Has(i) {
			ts.Fatalf(t, "SetRange(%v,%v) unexpected setted idx %v\n", from, to, i)
		}
	}
}

func testClearRange(b Bitmap, t *testing.T, set, from, to int) {

	b.ClearRange(0, 1024)
	b.Set(set)
	b.SetRange(from, to)

	b.ClearRange(from, to)

	for i := from; i <= to; i++ {
		if b.Has(i) {
			ts.Fatalf(t, "ClearRange(%v,%v) unexpected setted idx %v\n", from, to, i)
		}
	}
	if set >= 0 && !b.Has(set) {
		ts.Fatalf(t, "ClearRange(%v,%v) unexpected cleared idx %v\n", from, to, set)
	}
}

func testFindFromWithIndex(b Bitmap, t *testing.T, p int) {
	{
		b.ClearRange(0, p+1000000)
		b.SetRange(p+1, p+1000000)
		idx, _ := b.FindFrom(p, false)
		if idx != p+1 {
			t.Fatalf("b.FindFrom %v unexpected %v\n", p+1, idx)
		}
	}
	{
		b.ClearRange(0, p+1000000)
		b.Set(p + 1)
		idx, _ := b.FindFrom(p, false)
		if idx != p+1 {
			t.Fatalf("b.FindFrom %v unexpected %v\n", p+1, idx)
		}
	}
}

func testBigRange(b Bitmap, t *testing.T, step int) {

	b.ClearRange(0, step*3)
	for i := 0; i < 3; i++ {
		b.SetRange(i*step, (i+1)*step-1)
	}

	for i := 0; i < 3; i++ {
		idx, _ := b.FindFrom(i*step, false)
		if idx != i*step {
			t.Fatalf("b.FindFrom %v unexpected %v\n", i*step, idx)
		}
		b.Clear(idx)
	}

	for i := 0; i < 3; i++ {
		idx, _ := b.FindFrom(i*step, false)
		if idx != i*step+1 {
			t.Fatalf("b.FindFrom %v unexpected %v\n", i*step+1, idx)
		}
		b.Clear(idx)
	}
}

// -----------------------------------------------------------
