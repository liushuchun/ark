package crc32util

import (
	"bytes"
	"crypto/rand"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/qiniu/errors"
	"github.com/stretchr/testify.v1/assert"
	"github.com/stretchr/testify.v1/require"
)

func TestSize(t *testing.T) {
	fsizes := []int64{
		0,
		1,
		5,
		BufSize - 5,
		BufSize - 4,
		BufSize - 3,
		BufSize - 2,
		BufSize - 1,
		BufSize,
		BufSize + 1,
		BufSize + 2,
		2 * (BufSize - 5),
		2 * (BufSize - 4),
		2 * (BufSize - 3),
		2 * (BufSize - 2),
		2 * (BufSize - 1),
		2 * (BufSize),
		2 * (BufSize + 1),
		2 * (BufSize + 2),
	}
	for _, fsize := range fsizes {
		total := EncodeSize(fsize)
		fsize2 := DecodeSize(total)
		if fsize2 != fsize {
			t.Errorf("fsize: %v, totalSize: %v, fsize2: %v\n", fsize, total, fsize2)
		}
	}
}

func TestEncodeDecode(t *testing.T) {

	data := randData(128*1024 + 80)
	fsize := int64(len(data)) // 128k+80
	r := bytes.NewReader(data)
	w := bytes.NewBuffer(nil)
	err := Encode(w, r, fsize, nil)
	require.NoError(t, err)

	_64k := int64(BufSize)

	tcs := []struct {
		base  int64
		from  int64
		to    int64
		fsize int64
		tail  int64
	}{
		{from: 0, to: 0, fsize: 0},
		{from: fsize, to: fsize, fsize: fsize},
		{from: 0, to: fsize, fsize: fsize},

		{from: _64k - 4, to: fsize, fsize: fsize},
		{from: _64k + 0, to: fsize, fsize: fsize},
		{from: _64k + 4, to: fsize, fsize: fsize},
		{from: _64k + 5, to: fsize, fsize: fsize},

		{from: _64k - 4, to: fsize - 1, fsize: fsize},
		{from: _64k + 0, to: fsize - 1, fsize: fsize},
		{from: _64k + 4, to: fsize - 1, fsize: fsize},
		{from: _64k + 5, to: fsize - 1, fsize: fsize},

		{from: _64k + 4, to: fsize - _64k, fsize: fsize},
		{from: _64k + 4, to: fsize - _64k - 4, fsize: fsize},
		{from: _64k + 4, to: fsize - _64k - 5, fsize: fsize},
		{from: 0, to: fsize - _64k - 4 - _64k - 4, fsize: fsize},

		{base: 1, from: _64k + 4, to: fsize, fsize: fsize},
		{base: 0, from: _64k + 4, to: fsize, fsize: fsize, tail: 1000},
		{base: _64k + 4, from: _64k + 4, to: fsize, fsize: fsize, tail: _64k + 4},
		{base: _64k + 5, from: _64k + 4, to: fsize, fsize: fsize, tail: _64k + 5},
	}
	chunk := make([]byte, BufSize)
	for _, tc := range tcs {
		// log.Printf("%+v\n", tc)
		disk := append(make([]byte, tc.base), w.Bytes()...)
		disk = append(disk, make([]byte, tc.tail)...)

		dr := RangeDecoder(bytes.NewReader(disk), tc.base, chunk, tc.from, tc.to, tc.fsize)
		all, err := ioutil.ReadAll(dr)
		assert.NoError(t, err, "%+v", tc)
		assert.Equal(t, crc32.ChecksumIEEE(data[tc.from:tc.to]), crc32.ChecksumIEEE(all), "%+v", tc)
	}
}

func TestDecoderFail(t *testing.T) {
	data := bytes.Repeat([]byte("hellowor"), 2*1024*8+10)
	fsize := int64(len(data))
	r := bytes.NewReader(data)
	w := bytes.NewBuffer(nil)
	err := Encode(w, r, fsize, nil)
	assert.NoError(t, err)

	r2 := bytes.NewReader(w.Bytes())
	er := &errorReaderAt{r2, 3}
	dr := RangeDecoder(er, 0, nil, 0, 10, 10)
	all := make([]byte, fsize)
	_, err = io.ReadFull(dr, all)
	assert.Error(t, err)
}

func TestDecoderShort(t *testing.T) {
	data := make([]byte, chunkLen+1)
	r := bytes.NewReader(data)
	w := bytes.NewBuffer(nil)
	err := Encode(w, r, int64(len(data)), nil)
	assert.NoError(t, err)

	for off := 0; off <= 4; off++ {
		r2 := bytes.NewReader(w.Bytes()[:BufSize+off])
		dr := RangeDecoder(r2, 0, nil, 0, int64(len(data)), int64(len(data)))
		b, err := ioutil.ReadAll(dr)
		assert.Equal(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, data[:chunkLen], b)
	}

	r2 := bytes.NewReader(w.Bytes())
	dr := RangeDecoder(r2, 0, nil, 0, int64(len(data)), int64(len(data)))
	b, err := ioutil.ReadAll(dr)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, b)
}

func randData(size int64) []byte {

	b := make([]byte, size)
	rand.Read(b)
	return b
}

func TestAppend(t *testing.T) {

	dir := "./testappend/"
	os.MkdirAll(dir, 0777)
	defer os.RemoveAll(dir)

	f, _ := os.Create(dir + "1")

	size := int64(3*1024*1024 - 112123)
	data := randData(size)

	sizes := []int64{0, 50*1024 + 123, 64*1024 - 4, 64 * 1024, 134*1024 + 121, 320 * 1024, 1024*1024 + 123, size}
	for i := 1; i < len(sizes); i++ {
		osize := sizes[i-1]
		nsize := sizes[i]
		ndata := data[osize:nsize]
		r := bytes.NewReader(ndata)
		err := AppendEncode(f, 4112, osize, r, nsize-osize, nil)
		assert.NoError(t, err, "%v", i)

		all := make([]byte, nsize)
		dr := RangeDecoder(f, 4112, nil, 0, nsize, nsize)
		_, err = io.ReadFull(dr, all)
		assert.NoError(t, err, "%v", i)

		crc := crc32.ChecksumIEEE(data[:nsize])
		assert.Equal(t, crc, crc32.ChecksumIEEE(all), "%v", i)
	}

	f.Close()

	f, _ = os.Create(dir + "2")

	size0 := int64(3*1024 + 1)
	data0 := randData(size0)

	var r io.Reader
	r = bytes.NewReader(data0)
	err := AppendEncode(f, 123, 0, r, size0, nil)
	assert.NoError(t, err)

	crc := crc32.ChecksumIEEE(data0)
	all := make([]byte, size0)
	dr := RangeDecoder(f, 123, nil, 0, size0, size0)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	size1 := int64(23*1024 + 13)
	data1 := randData(size1)

	r = &errorReader{bytes.NewReader(data1), 20*1024 + 12}
	err = AppendEncode(f, 123, size0, r, size1, nil)
	assert.Error(t, err)

	all = make([]byte, size0)
	dr = RangeDecoder(f, 123, nil, 0, size0, size0)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	r = bytes.NewReader(data1)
	err = AppendEncode(f, 123, size0, r, size1, nil)
	assert.NoError(t, err)
	crc = crc32.Update(crc, crc32.IEEETable, data1)

	all = make([]byte, size0+size1)
	dr = RangeDecoder(f, 123, nil, 0, size0+size1, size0+size1)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	size2 := int64(123*1024 + 313)
	data2 := randData(size2)

	r = &errorReader{bytes.NewReader(data2), 100*1024 + 436}
	err = AppendEncode(f, 123, size0+size1, r, size2, nil)
	assert.Error(t, err)

	all = make([]byte, size0+size1)
	dr = RangeDecoder(f, 123, nil, 0, size0+size1, size0+size1)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	r = bytes.NewReader(data2)
	err = AppendEncode(f, 123, size0+size1, r, size2, nil)
	assert.NoError(t, err)
	crc = crc32.Update(crc, crc32.IEEETable, data2)

	all = make([]byte, size0+size1+size2)
	dr = RangeDecoder(f, 123, nil, 0, size0+size1+size2, size0+size1+size2)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	f.Close()
}

type errorReader struct {
	io.Reader
	remain int
}

func (p *errorReader) Read(b []byte) (int, error) {

	if p.remain == 0 {
		return 0, errors.New("errorReader: hit")
	}
	if len(b) > p.remain {
		b = b[:p.remain]
	}
	n, err := p.Reader.Read(b)
	p.remain -= n
	return n, err
}

type errorReaderAt struct {
	io.ReaderAt
	remain int
}

func (p *errorReaderAt) ReadAt(b []byte, off int64) (int, error) {

	if p.remain == 0 {
		return 0, errors.New("errorReader: hit")
	}
	if len(b) > p.remain {
		b = b[:p.remain]
	}
	n, err := p.ReaderAt.ReadAt(b, off)
	p.remain -= n
	return n, err
}
