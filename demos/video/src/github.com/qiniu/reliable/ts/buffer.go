package ts

import (
	"github.com/qiniu/bytes"
	"os"
	"syscall"
	"time"
)

// --------------------------------------------------------------------

var zeroTime time.Time

type fileInfo struct {
	fsize int64
}

func (p *fileInfo) Name() string {
	return ""
}

func (p *fileInfo) Size() int64 {
	return p.fsize
}

func (p *fileInfo) Mode() os.FileMode {
	return 0666
}

func (p *fileInfo) ModTime() time.Time {
	return zeroTime
}

func (p *fileInfo) IsDir() bool {
	return false
}

func (p *fileInfo) Sys() interface{} {
	return nil
}

// --------------------------------------------------------------------

const (
	WriteOk = iota
	WriteFail
	WriteBad
	WriteShort
)

type Buffer struct {
	*bytes.Buffer
	modes []int
}

func NewBuffer(modes []int) *Buffer {

	b := bytes.NewBuffer()
	return &Buffer{b, modes}
}

func (p *Buffer) Close() error {

	return nil
}

func (p *Buffer) Stat() (fi os.FileInfo, err error) {

	fsize := int64(p.Len())
	return &fileInfo{fsize: fsize}, nil
}

func (p *Buffer) WriteAt(buf []byte, off int64) (n int, err error) {

	var mode int
	if len(p.modes) > 0 {
		mode = p.modes[0]
		p.modes = p.modes[1:]
	}
	switch mode {
	case WriteShort:
		nl := len(buf) - 5
		if nl < 0 {
			nl = 0
		}
		buf = buf[:nl]
		fallthrough
	case WriteOk:
		return p.Buffer.WriteAt(buf, off)
	case WriteFail:
		return 0, syscall.EFAULT
	}
	b := make([]byte, len(buf))
	copy(b, buf)
	if len(b) >= 12 {
		b[10] ^= 0x8F
		b[11] ^= 0xF7
	} else {
		b[0] ^= 0x77
	}
	return p.Buffer.WriteAt(b, off)
}

// --------------------------------------------------------------------
