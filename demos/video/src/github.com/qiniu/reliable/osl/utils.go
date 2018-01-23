package osl

import (
	"github.com/qiniu/log.v1"
	"github.com/qiniu/reliable/errors"
	"os"
	"time"
)

var ErrTooManyFails = errors.ErrTooManyFails

// --------------------------------------------------------------------

type File interface {
	ReadAt(buf []byte, off int64) (n int, err error)
	WriteAt(buf []byte, off int64) (n int, err error)
	Truncate(fsize int64) (err error)
	Stat() (fi os.FileInfo, err error)
	Close() (err error)
}

func Open(fnames []string, allowfails int) (files []File, err error) {

	fails := 0
	files = make([]File, len(fnames))
	for i, fname := range fnames {
		f2, err2 := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0666)
		if err2 != nil {
			log.Warn("reliable.Open: os.Open failed -", err2)
			fails++
			if fails > allowfails {
				for j := 0; j < i; j++ {
					if files[j] != nil {
						files[j].Close()
						files[j] = nil
					}
				}
				return nil, ErrTooManyFails
			}
			continue
		}
		files[i] = f2
	}
	return
}

func FsizeOf(files []File, allowfails int) (fsize int64, err error) {

	fails := 0
	for _, f := range files {
		if f != nil {
			fi2, err2 := f.Stat()
			if err2 == nil {
				fsize2 := fi2.Size()
				if fsize < fsize2 {
					fsize = fsize2
				}
				continue
			}
			log.Warn("FsizeOf: f.Stat failed -", err2)
		}
		fails++
		if fails > allowfails {
			log.Error("FsizeOf failed: too many fails")
			return 0, ErrTooManyFails
		}
	}
	return
}

// --------------------------------------------------------------------

var zeroTime time.Time

type FileInfo struct {
	Fsize int64
}

func (p *FileInfo) Name() string {
	return ""
}

func (p *FileInfo) Size() int64 {
	return p.Fsize
}

func (p *FileInfo) Mode() os.FileMode {
	return 0666
}

func (p *FileInfo) ModTime() time.Time {
	return zeroTime
}

func (p *FileInfo) IsDir() bool {
	return false
}

func (p *FileInfo) Sys() interface{} {
	return nil
}

// --------------------------------------------------------------------
