package crc32util

import (
	"encoding/binary"
	"hash/crc32"
	"io"

	"github.com/qiniu/errors"
	qio "github.com/qiniu/io"
)

const (
	chunkBits = 16 // 64K
	chunkLen  = (1 << chunkBits) - 4
)

const (
	BufSize = chunkLen + 4
)

var (
	ErrUnmatchedChecksum = errors.New("unmatched checksum")
)

// -----------------------------------------

func EncodeSize(fsize int64) int64 {

	chunkCount := (fsize + (chunkLen - 1)) / chunkLen
	return fsize + 4*chunkCount
}

func DecodeSize(totalSize int64) int64 {

	chunkCount := (totalSize + (BufSize - 1)) / BufSize
	return totalSize - 4*chunkCount
}

// -----------------------------------------

type ReaderError struct {
	error
}

type WriterError struct {
	error
}

func Encode(w io.Writer, in io.Reader, fsize int64, chunk []byte) (err error) {

	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Encode failed: invalid len(chunk)")
	}

	i := 0
	for fsize >= chunkLen {
		_, err = io.ReadFull(in, chunk[4:])
		if err != nil {
			return ReaderError{err}
		}
		crc := crc32.ChecksumIEEE(chunk[4:])
		binary.LittleEndian.PutUint32(chunk, crc)
		_, err = w.Write(chunk)
		if err != nil {
			return WriterError{err}
		}
		fsize -= chunkLen
		i++
	}

	if fsize > 0 {
		n := fsize + 4
		_, err = io.ReadFull(in, chunk[4:n])
		if err != nil {
			return ReaderError{err}
		}
		crc := crc32.ChecksumIEEE(chunk[4:n])
		binary.LittleEndian.PutUint32(chunk, crc)
		_, err = w.Write(chunk[:n])
		if err != nil {
			err = WriterError{err}
		}
	}
	return
}

// ---------------------------------------------------------------------------

type ReaderWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

//
// 对于数据开始在(rw io.ReaderWriterAt, base int64) 的文件大小为 fsize 的做 crc32 冗余校验的文件
// 我们要往它后面追加 size 大小的数据
func AppendEncode(rw ReaderWriterAt, base int64, fsize int64, in io.Reader, size int64, chunk []byte) (err error) {

	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Encode failed: invalid len(chunk)")
	}

	offset := base + EncodeSize(fsize)
	if oldSize := fsize % chunkLen; oldSize > 0 {
		// 旧文件的最后一个 chunk 需要特殊处理。
		// 处理流程为：读取旧内容、写入新内容、写入总 crc32。
		r := RangeDecoder(rw, base, chunk, fsize-oldSize, fsize, fsize)
		_, err = io.ReadFull(r, chunk[4:4+oldSize])
		if err != nil {
			// 从 rw 读失败，认为是 writer 错误。
			return WriterError{err}
		}
		addSize := chunkLen - oldSize
		if addSize > size {
			addSize = size
		}
		add := chunk[4+oldSize : 4+oldSize+addSize]
		_, err = io.ReadFull(in, add)
		if err != nil {
			return ReaderError{err}
		}
		// 如果 header 写成功但 data 写失败，这个 chunk 就无法正常读写了。
		// 因此下面的操作是先写 data 再写 header。
		_, err = rw.WriteAt(add, offset)
		if err != nil {
			return WriterError{err}
		}
		crc := crc32.ChecksumIEEE(chunk[4 : 4+oldSize+addSize])
		pos := base + (fsize/chunkLen)<<chunkBits
		defer func() {
			if err != nil {
				return
			}
			binary.LittleEndian.PutUint32(chunk[:4], crc)
			_, err = rw.WriteAt(chunk[:4], pos)
			if err != nil {
				err = WriterError{err}
			}
		}()
		size -= addSize
		offset += addSize
	}
	if size == 0 {
		return nil
	}
	w := &qio.Writer{
		WriterAt: rw,
		Offset:   offset,
	}
	return Encode(w, in, size, chunk)
}

// ---------------------------------------------------------------------------

type decoder struct { // raw+crc32 input => raw input
	chunk   []byte
	in      io.Reader
	lastErr error
	off     int
	left    int64
}

func Decoder(in io.Reader, n int64, chunk []byte) (dec *decoder) {

	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Decoder failed: invalid len(chunk)")
	}

	dec = &decoder{chunk, in, nil, BufSize, n}
	return
}

func (r *decoder) fetch() {

	min := len(r.chunk)
	if r.left+4 < int64(min) {
		min = int(r.left + 4)
	}
	var n2 int
	n2, r.lastErr = io.ReadAtLeast(r.in, r.chunk, min)
	if r.lastErr != nil {
		if r.lastErr == io.EOF {
			r.lastErr = io.ErrUnexpectedEOF
		}
		return
	}
	crc := crc32.ChecksumIEEE(r.chunk[4:n2])
	if binary.LittleEndian.Uint32(r.chunk) != crc {
		r.lastErr = errors.Info(ErrUnmatchedChecksum, "crc32util.decode")
		return
	}
	r.chunk = r.chunk[:n2]
	r.off = 4
}

func (r *decoder) Read(b []byte) (n int, err error) {

	if r.off == len(r.chunk) {
		if r.lastErr != nil {
			err = r.lastErr
			return
		}
		if r.left == 0 {
			err = io.EOF
			return
		}
		r.fetch()
	}
	n = copy(b, r.chunk[r.off:])
	r.off += n
	r.left -= int64(n)
	return
}

// ---------------------------------------------------------------------------

//
// 对于数据开始在 (in io.ReaderAt, base int64)，文件大小为 fsize 的做 crc32 冗余校验的文件，我们
// 要读取其中 [from, to) 范围的数据。
//
func RangeDecoder(in io.ReaderAt, base int64, chunk []byte, from, to, fsize int64) io.Reader {

	fromBase := (from / chunkLen) << chunkBits
	encodedSize := EncodeSize(fsize) - fromBase
	sect := io.NewSectionReader(in, base+fromBase, encodedSize)
	dec := Decoder(sect, DecodeSize(encodedSize), chunk)
	if (from == 0 || from%chunkLen == 0) && to >= fsize {
		return dec
	}
	return newSectionReader(dec, from%chunkLen, to-from)
}

// ---------------------------------------------------------------------------

func decodeAt(w io.Writer, in io.ReaderAt, chunk []byte, idx int64, ifrom, ito int) (err error) {

	n, err := in.ReadAt(chunk, idx<<chunkBits)
	if err != nil {
		if err != io.EOF {
			return
		}
	}
	if n <= 4 {
		if n == 0 {
			return io.EOF
		}
		err = errors.Info(io.ErrUnexpectedEOF, "crc32util.Decode", "n:", n)
		return
	}

	crc := crc32.ChecksumIEEE(chunk[4:n])
	if binary.LittleEndian.Uint32(chunk) != crc {
		err = errors.Info(ErrUnmatchedChecksum, "crc32util.Decode")
		return
	}

	ifrom += 4
	ito += 4
	if ito > n {
		ito = n
	}
	if ifrom >= ito {
		return io.EOF
	}
	_, err = w.Write(chunk[ifrom:ito])
	return
}

func DecodeRange(w io.Writer, in io.ReaderAt, chunk []byte, from, to int64) (err error) {

	if from >= to {
		return
	}

	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Decode failed: invalid len(chunk)")
	}

	fromIdx, toIdx := from/chunkLen, to/chunkLen
	fromOff, toOff := int(from%chunkLen), int(to%chunkLen)
	if fromIdx == toIdx { // 只有一行
		return decodeAt(w, in, chunk, fromIdx, fromOff, toOff)
	}
	for fromIdx < toIdx {
		err = decodeAt(w, in, chunk, fromIdx, fromOff, chunkLen)
		if err != nil {
			return
		}
		fromIdx++
		fromOff = 0
	}
	if toOff > 0 {
		err = decodeAt(w, in, chunk, fromIdx, 0, toOff)
	}
	return
}

// ---------------------------------------------------------------------------
