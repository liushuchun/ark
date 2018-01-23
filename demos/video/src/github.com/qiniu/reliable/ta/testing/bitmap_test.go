package ta

import (
	"github.com/qiniu/reliable/bits"
	. "github.com/qiniu/reliable/ta"
	"os"
	"strconv"
	"testing"
)

type bitmapTester struct {
	bitmap IBitmap
	fname  string
}

func createBitmapTester(fname string, id int, ta *Transaction) taTester {
	bitmap1, err := bits.OpenBitfile([]string{fname}, 8, 1)
	if err != nil {
		panic(err)
	}
	bitmap2, err := OpenBitmap(bitmap1, ta, id)
	if err != nil {
		panic(err)
	}
	return &bitmapTester{bitmap2, fname}
}

func createBigBitmapTester(fname string, id int, ta *Transaction) taTester {
	bitmap1, err := bits.OpenBigBitfile([]string{fname}, 8, 100000000, 1)
	if err != nil {
		panic(err)
	}
	bitmap2, err := OpenBitmap(bitmap1, ta, id)
	if err != nil {
		panic(err)
	}
	return &bitmapTester{bitmap2, fname}
}

func (t *bitmapTester) clear() {
	os.Remove(t.fname)
}

func (t *bitmapTester) setA(mid int) {
	err := t.bitmap.Set(mid)
	if err != nil {
		panic(err)
	}
}

func (t *bitmapTester) setB(mid int) {
	err := t.bitmap.Clear(mid)
	if err != nil {
		panic(err)
	}
}

func (t *bitmapTester) checkA(mid int) {
	if !t.bitmap.Has(mid) {
		panic("[" + strconv.Itoa(mid) + "] got: false, should be: true")
	}
}

func (t *bitmapTester) checkB(mid int) {
	if t.bitmap.Has(mid) {
		panic("[" + strconv.Itoa(mid) + "] got: true, should be: false")
	}
}

func TestBitmap(t *testing.T) {
	fname1 := "test_bitmap.qboxtest"
	fname2 := "test_bigbitmap.qboxtest"
	defer os.Remove(fname1)
	defer os.Remove(fname2)
	testTaTester(65, func(ta *Transaction) taTester {
		return createBitmapTester(fname1, 0, ta)
	})
	testTaTester(65, func(ta *Transaction) taTester {
		return createBigBitmapTester(fname2, 0, ta)
	})
}
