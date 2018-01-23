package bufio

import (
	"io"
)

const (
	defaultReaderAtRoundBits = 16      // 64k
	defaultReaderAtBufSize   = 1 << 20 // 1MB
)

// ---------------------------------------------------

type ReaderAt struct {
	f         io.ReaderAt
	buffer    []byte
	base      int64
	roundMask int64
	n         int
}

func NewReaderAt(f io.ReaderAt) *ReaderAt {
	return NewReaderAtSize(f, make([]byte, defaultReaderAtBufSize), defaultReaderAtRoundBits)
}

func NewReaderAtSize(f io.ReaderAt, buffer []byte, roundBits uint) *ReaderAt {

	roundMask := (1 << roundBits) - 1
	if (len(buffer) & roundMask) != 0 {
		panic("len(buffer) is not round") // gocov: rare
	}

	return &ReaderAt{
		f: f, buffer: buffer, roundMask: int64(roundMask),
	}
}

// ---------------------------------------------------

func (p *ReaderAt) ReadAt(buf []byte, off int64) (n int, err error) {

	var ioff, n1 int

	if off < p.base || off >= p.base+int64(p.n) {
		goto lzRead
	}
	ioff = int(off - p.base)

lzRetData: // gocov: 误报
	n1 = copy(buf, p.buffer[ioff:p.n])
	n += n1
	if len(buf) == n1 {
		return n, nil
	}
	if p.n != len(p.buffer) {
		err = io.EOF
		return
	}
	buf = buf[n1:]
	off += int64(n1)

lzRead:
	p.base = off &^ p.roundMask
	p.n, err = p.f.ReadAt(p.buffer, p.base)
	if err != nil {
		if err != io.EOF {
			return // gocov: rare
		}
	}
	ioff = int(off - p.base)
	if ioff >= p.n {
		err = io.EOF
		return
	}
	goto lzRetData
}

// ---------------------------------------------------
