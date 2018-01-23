package bufio

import (
	"bytes"
	"io"
	"testing"
)

// ---------------------------------------------------

func newReader() io.ReaderAt {

	b := make([]byte, 253)
	for i := range b {
		b[i] = byte(i)
	}
	return bytes.NewReader(b)
}

func doTest(t *testing.T, a io.ReaderAt, b *ReaderAt, off int64, n int) {

	b1 := make([]byte, n)
	b2 := make([]byte, n)

	n1, err1 := a.ReadAt(b1, off)
	n2, err2 := b.ReadAt(b2, off)

	if n1 != n2 || !bytes.Equal(b1[:n1], b2[:n2]) {
		t.Fatal("n1 != n2 || !bytes.Equal(b1[:n1], b2[:n2])")
	}

	if err1 != err2 {
		if (err1 == nil || err1 == io.EOF) || (err2 == nil || err2 == io.EOF) {
			t.Fatal("err1 != err2 -", err1, err2)
		}
	}
}

func Test(t *testing.T) {

	a := newReader()
	b := NewReaderAtSize(a, make([]byte, 64), 4)
	doTest(t, a, b, 0, 3)
	doTest(t, a, b, 73, 7)
	doTest(t, a, b, 255, 7)
	doTest(t, a, b, 3, 255)
	doTest(t, a, b, 3, 81)
	doTest(t, a, b, 93, 255)
	doTest(t, a, b, 83, 10)

	b = NewReaderAt(a)
	doTest(t, a, b, 0, 3)
	doTest(t, a, b, 73, 7)
	doTest(t, a, b, 255, 7)
	doTest(t, a, b, 3, 255)
	doTest(t, a, b, 3, 81)
	doTest(t, a, b, 93, 255)
	doTest(t, a, b, 83, 10)
}

// ---------------------------------------------------
