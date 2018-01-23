package singletrip

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var pollInterval = 100 * time.Millisecond

var (
	errWriteFinished = errors.New("singletrip.v2: write is finished")
	errWaitTimeout   = errors.New("singletrip.v2: wait timeout")
)

type buffer interface {
	Write(p []byte) (n int, err error)
	ReadAt(p []byte, off int64) (n int, err error)
	WaitWrite(off int64, readTimeout time.Duration) error
	FinishWrite(err error)
	WriteFinished() (bool, error)
	Len() int
	Close() error
}

type byteBuffer struct {
	mu       sync.Mutex
	b        []byte // TODO: use pool.
	err      error
	finished bool
}

func (buf *byteBuffer) Write(p []byte) (n int, err error) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	if buf.finished {
		return 0, errWriteFinished
	}
	buf.b = append(buf.b, p...)
	return len(p), nil
}

func (buf *byteBuffer) ReadAt(p []byte, off int64) (n int, err error) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	ioff := int(off)
	if buf.b == nil || ioff >= len(buf.b) {
		return 0, io.EOF
	}
	n = copy(p, buf.b[ioff:])
	if n != len(p) {
		err = io.EOF
	}
	return
}

func (buf *byteBuffer) WaitWrite(off int64, readTimeout time.Duration) error {
	ioff := int(off)
	start := time.Now()
	for {
		buf.mu.Lock()
		if buf.finished || len(buf.b) > ioff {
			buf.mu.Unlock()
			return nil
		}
		if readTimeout > 0 && time.Since(start) > readTimeout {
			buf.mu.Unlock()
			return errWaitTimeout
		}
		buf.mu.Unlock()
		time.Sleep(pollInterval)
	}
}

func (buf *byteBuffer) FinishWrite(err error) {
	buf.mu.Lock()
	buf.finished, buf.err = true, err
	buf.mu.Unlock()
}

func (buf *byteBuffer) WriteFinished() (bool, error) {
	buf.mu.Lock()
	finished, err := buf.finished, buf.err
	buf.mu.Unlock()
	return finished, err
}

func (buf *byteBuffer) Len() int {
	buf.mu.Lock()
	n := len(buf.b)
	buf.mu.Unlock()
	return n
}

func (buf *byteBuffer) Close() error {
	return nil
}

type fileBuffer struct {
	mu       sync.Mutex
	f        *os.File
	woff     int
	err      error
	finished bool
}

func (buf *fileBuffer) Write(p []byte) (n int, err error) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	if buf.finished {
		return 0, errWriteFinished
	}
	n, err = buf.f.Write(p)
	if err != nil {
		buf.setErr(err)
	}
	buf.woff += n
	return
}

func (buf *fileBuffer) setErr(err error) {
	buf.finished, buf.err = true, err
}

func (buf *fileBuffer) ReadAt(p []byte, off int64) (n int, err error) {
	buf.mu.Lock()
	n, err = buf.f.ReadAt(p, off)
	buf.mu.Unlock()
	return
}

func (buf *fileBuffer) WaitWrite(off int64, readTimeout time.Duration) error {
	ioff := int(off)
	start := time.Now()
	for {
		buf.mu.Lock()
		if buf.finished || buf.woff > ioff {
			buf.mu.Unlock()
			return nil
		}
		if readTimeout > 0 && time.Since(start) > readTimeout {
			buf.mu.Unlock()
			return errWaitTimeout
		}
		buf.mu.Unlock()
		time.Sleep(pollInterval)
	}
}

func (buf *fileBuffer) FinishWrite(err error) {
	buf.mu.Lock()
	buf.setErr(err)
	buf.mu.Unlock()
}

func (buf *fileBuffer) WriteFinished() (bool, error) {
	buf.mu.Lock()
	finished, err := buf.finished, buf.err
	buf.mu.Unlock()
	return finished, err
}

func (buf *fileBuffer) Len() int {
	buf.mu.Lock()
	n := buf.woff
	buf.mu.Unlock()
	return n
}

func (buf *fileBuffer) Close() error {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	var errs []string
	if err := buf.f.Close(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := os.Remove(buf.f.Name()); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

type bufferReader struct {
	buf         buffer
	off         int64
	readTimeout time.Duration
	closeFn     func()
}

// 此函数仅用于单元测试。
var afterReadAtEOFHook func()

func (r *bufferReader) Read(p []byte) (n int, err error) {
	for {
		n, err = r.buf.ReadAt(p, r.off)
		r.off += int64(n)
		if err == nil {
			return
		}
		if err != io.EOF {
			return
		}
		if afterReadAtEOFHook != nil {
			afterReadAtEOFHook()
		}
		// err == io.EOF
		if ok, ierr := r.buf.WriteFinished(); ok {
			if int(r.off) >= r.buf.Len() {
				// 读到 buffer 尾部，返回。
				if ierr != nil {
					err = ierr
				}
				return
			}
			// 虽然写完成了，但是还没有读取到 buffer 尾部，重新读取。
			r.off -= int64(n)
			continue
		}
		if err = r.buf.WaitWrite(r.off, r.readTimeout); err != nil {
			return
		}
		// 有新的写入，重新读取。
		r.off -= int64(n)
	}
}

func (r *bufferReader) Close() error {
	if r.closeFn != nil {
		r.closeFn()
	}
	return nil
}
