package ta

import (
	"bytes"
	"github.com/qiniu/reliable"
	. "github.com/qiniu/reliable/ta"
	"os"
	"strconv"
	"testing"
)

type arrayTester struct {
	array IArray
	fname string
}

func createArrayTester(fname string, id int, ta *Transaction) taTester {
	array1, err := reliable.OpenArrfile([]string{fname}, 8, 1)
	if err != nil {
		panic(err)
	}
	array2, err := OpenArray(array1, ta, id)
	if err != nil {
		panic(err)
	}
	return &arrayTester{array2, fname}
}

func createBigArrayTester(fname string, id int, ta *Transaction) taTester {
	array1, err := reliable.OpenBigArrfile([]string{fname}, 8, 8, 1)
	if err != nil {
		panic(err)
	}
	array2, err := OpenArray(array1, ta, id)
	if err != nil {
		panic(err)
	}
	return &arrayTester{array2, fname}
}

func (t *arrayTester) clear() {
	os.Remove(t.fname)
}

func (t *arrayTester) set(mid int, val string) {
	buf := make([]byte, 8)
	copy(buf, []byte(val))
	err := t.array.Put(mid, buf)
	if err != nil {
		panic(err)
	}
}

func (t *arrayTester) check(mid int, val string) {
	buf := make([]byte, 8)
	err := t.array.Get(mid, buf)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(buf, []byte(val)) {
		panic("[" + strconv.Itoa(mid) + "] got: " + string(buf) + ", should be: " + val)
	}
}

func (t *arrayTester) setA(mid int) {
	t.set(mid, "abcdefgh")
}

func (t *arrayTester) setB(mid int) {
	t.set(mid, "ABCDEFGH")
}

func (t *arrayTester) checkA(mid int) {
	t.check(mid, "abcdefgh")
}

func (t *arrayTester) checkB(mid int) {
	t.check(mid, "ABCDEFGH")
}

func TestArray(t *testing.T) {
	fname1 := "test_array.qboxtest"
	fname2 := "test_bigarray.qboxtest"
	defer os.Remove(fname1)
	defer os.Remove(fname2)
	testTaTester(65, func(ta *Transaction) taTester {
		return createArrayTester(fname1, 0, ta)
	})
	testTaTester(65, func(ta *Transaction) taTester {
		return createBigArrayTester(fname2, 0, ta)
	})
}
