package fixed

import (
	"os"
	"strconv"
	"syscall"

	"github.com/qiniu/errors"
)

var ErrOutOfRange = errors.New("i/o error: out of file range")

// --------------------------------------------------------------------

const (
	DefaultChunkBits = 26
)

type File struct {
	files     []*os.File
	fsize     int64
	chunkBits uint
}

func Open(name string, chunkBits uint, fsize int64) (r *File, err error) {

	err = syscall.Mkdir(name, 0777)
	if err != nil {
		if err != syscall.EEXIST {
			err = errors.Info(err, "largefile.Open failed", name).Detail(err)
			return
		}
		err = nil
	}

	if chunkBits > 32 {
		err = errors.Info(syscall.EINVAL, "largefile.Open failed: invalid argument")
		return
	} else if chunkBits == 0 {
		chunkBits = DefaultChunkBits
	}

	base := name + "/"
	n := int(fsize >> chunkBits)
	files := make([]*os.File, n)
	r = &File{files, fsize, chunkBits}

	for idx := 0; idx < n; idx++ {
		fp, err2 := os.OpenFile(base+strconv.FormatInt(int64(idx), 36), os.O_RDWR|os.O_CREATE, 0666)
		if err2 != nil {
			err = errors.Info(err2, "largefile.Open failed: os.OpenFile", idx).Detail(err2)
			r.Close()
			return
		}
		files[idx] = fp
	}
	return
}

func (r *File) Close() (err error) {

	for idx, f := range r.files {
		if f != nil {
			f.Close()
			r.files[idx] = nil
		}
	}
	return nil
}

func (r *File) getFile(off int64) (f *os.File, offNew int64, sizeLeft int, err error) {

	if off > r.fsize {
		err = ErrOutOfRange
		return
	}

	chunkBits := r.chunkBits
	idx := off >> chunkBits

	f = r.files[idx]
	offNew = off - (idx << chunkBits)
	sizeLeft = (1 << chunkBits) - int(offNew)
	return
}

func (r *File) ReadAt(buf []byte, off int64) (n int, err error) {

	f, offNew, sizeLeft, err := r.getFile(off)
	if err != nil {
		err = errors.Info(err, "largefile.File.ReadAt failed").Detail(err)
		return
	}

	if len(buf) <= sizeLeft {
		return f.ReadAt(buf, offNew)
	}

	n, err = f.ReadAt(buf[:sizeLeft], offNew)
	if err != nil {
		return
	}

	n2, err := r.ReadAt(buf[sizeLeft:], off+int64(sizeLeft))
	n += n2
	return
}

func (r *File) WriteAt(buf []byte, off int64) (n int, err error) {

	f, offNew, sizeLeft, err := r.getFile(off)
	if err != nil {
		err = errors.Info(err, "largefile.File.WriteAt failed").Detail(err)
		return
	}

	if len(buf) <= sizeLeft {
		return f.WriteAt(buf, offNew)
	}

	n, err = f.WriteAt(buf[:sizeLeft], offNew)
	if err != nil {
		return
	}

	n2, err := r.WriteAt(buf[sizeLeft:], off+int64(sizeLeft))
	n += n2
	return
}

// --------------------------------------------------------------------
