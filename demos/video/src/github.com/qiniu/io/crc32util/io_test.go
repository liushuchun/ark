package crc32util

import (
	"io"
	"strings"
	"testing"
)

func TestSectionReader_ReadAt(t *testing.T) {
	dat := "a long sample data, 1234567890"
	tests := []struct {
		data   string
		off    int
		n      int
		bufLen int
		exp    string
		err    error
	}{
		{data: "", off: 0, n: 10, bufLen: 2, exp: "", err: io.EOF},
		{data: dat, off: 0, n: len(dat), bufLen: 0, exp: "", err: nil},
		{data: dat, off: len(dat), n: 1, bufLen: 1, exp: "", err: io.EOF},
		{data: dat, off: 0, n: len(dat) + 2, bufLen: len(dat), exp: dat, err: nil},
		{data: dat, off: 0, n: len(dat), bufLen: len(dat) / 2, exp: dat[:len(dat)/2], err: nil},
		{data: dat, off: 0, n: len(dat), bufLen: len(dat), exp: dat, err: nil},
		{data: dat, off: 2, n: len(dat), bufLen: len(dat) / 2, exp: dat[2 : 2+len(dat)/2], err: nil},
	}
	for i, tt := range tests {
		r := strings.NewReader(tt.data)
		s := newSectionReader(r, int64(tt.off), int64(tt.n))
		buf := make([]byte, tt.bufLen)
		if n, err := io.ReadFull(s, buf); n != len(tt.exp) || string(buf[:n]) != tt.exp || err != tt.err {
			t.Fatalf("%d: io.ReadFull = %q, %v; expected %q, %v", i, buf[:n], err, tt.exp, tt.err)
		}
	}
}
