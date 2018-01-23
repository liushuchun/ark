package singletrip

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"qiniupkg.com/x/log.v7"

	"github.com/stretchr/testify.v2/require"
)

var (
	tempDir  = "./run"
	tempDirs = []string{"./run/temp1", "./run/temp2", "./run/temp3"}
)

func TestMain(m *testing.M) {
	os.MkdirAll(tempDir, 0700)
	code := m.Run()
	os.RemoveAll(tempDir)
	os.Exit(code)
}

func TestBuffer(t *testing.T) {
	fn := func(buf buffer) {
		log.Infof("who: %T", buf)

		b := make([]byte, 20)
		n, err := buf.ReadAt(b, 0)
		require.Equal(t, io.EOF, err)
		require.Equal(t, 0, n)

		n, err = buf.Write([]byte("hel"))
		require.NoError(t, err)
		require.Equal(t, 3, n)
		n, err = buf.Write([]byte("loworld"))
		require.NoError(t, err)
		require.Equal(t, 7, n)

		n, err = buf.ReadAt(b, 0)
		require.Equal(t, io.EOF, err)
		require.Equal(t, "helloworld", string(b[:n]))

		err = buf.WaitWrite(10, 100*time.Millisecond)
		require.Equal(t, errWaitTimeout, err)
		err = buf.WaitWrite(2, 100*time.Millisecond)
		require.NoError(t, err)

		ferr := errors.New("finish write error")
		buf.FinishWrite(ferr)
		finished, err := buf.WriteFinished()
		require.True(t, finished)
		require.Equal(t, ferr, err)
		_, err = buf.Write([]byte{})
		require.Equal(t, errWriteFinished, err)
		err = buf.WaitWrite(10, 100*time.Millisecond)
		require.NoError(t, err)

		require.NoError(t, buf.Close())
	}

	bbuf := &byteBuffer{}
	fn(bbuf)

	f, err := ioutil.TempFile(tempDir, "singletrip")
	require.NoError(t, err)
	fbuf := &fileBuffer{f: f}
	fn(fbuf)
}

func TestFileBuffer_FileError(t *testing.T) {
	f, err := ioutil.TempFile(tempDir, "singletrip")
	require.NoError(t, err)
	buf := fileBuffer{f: f}
	f.Close()

	_, err = buf.Write([]byte("hello"))
	require.Error(t, err)
	finished, err := buf.WriteFinished()
	require.True(t, finished)
	require.Error(t, err)

	// close error: f is cloesd.
	f, err = ioutil.TempFile(tempDir, "singletrip")
	require.NoError(t, err)
	buf = fileBuffer{f: f}
	f.Close()
	err = buf.Close()
	require.Error(t, err)

	// close error: file is removed.
	f, err = ioutil.TempFile(tempDir, "singletrip")
	require.NoError(t, err)
	buf = fileBuffer{f: f}
	os.Remove(f.Name())
	err = buf.Close()
	require.Error(t, err)
}

func TestBufferReader_Error(t *testing.T) {
	f, err := ioutil.TempFile(tempDir, "singletrip")
	require.NoError(t, err)
	buf := &fileBuffer{f: f}
	br := &bufferReader{buf: buf}
	f.Close()
	b := make([]byte, 5)
	_, err = br.Read(b)
	require.Error(t, err)
	br.Close()
}

func TestBufferReader_Internal(t *testing.T) {
	defer func() {
		afterReadAtEOFHook = nil
	}()

	buf := &byteBuffer{}
	br := &bufferReader{buf: buf}

	buf.Write([]byte("hello"))
	afterReadAtEOFHook = func() {
		buf.Write([]byte("world"))
		buf.FinishWrite(nil)
	}
	b := make([]byte, 15)
	n, err := br.Read(b)
	require.Equal(t, io.EOF, err)
	require.Equal(t, "helloworld", string(b[:n]))

	buf = &byteBuffer{}
	br = &bufferReader{buf: buf}

	ferr := errors.New("finish write error")
	buf.Write([]byte("hello"))
	afterReadAtEOFHook = func() {
		buf.Write([]byte("world"))
		buf.FinishWrite(ferr)
	}
	b = make([]byte, 15)
	n, err = br.Read(b)
	require.Equal(t, ferr, err)
	require.Equal(t, "helloworld", string(b[:n]))

	buf = &byteBuffer{}
	br = &bufferReader{buf: buf}
	afterReadAtEOFHook = nil

	go func() {
		b = make([]byte, 15)
		n, err = br.Read(b)
		require.Equal(t, io.EOF, err)
		require.Equal(t, "helloworld", string(b[:n]))
	}()
	time.Sleep(100 * time.Millisecond)
	buf.Write([]byte("helloworld"))
	buf.FinishWrite(nil)
}

func TestBufferReader_Normal(t *testing.T) {
	buf := &byteBuffer{}
	br := &bufferReader{buf: buf, readTimeout: 100 * time.Millisecond}
	buf.Write([]byte("helloworld"))

	b := make([]byte, 5)
	n, err := br.Read(b)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, "hello", string(b))

	b = make([]byte, 10)
	n, err = br.Read(b)
	require.Equal(t, errWaitTimeout, err)
	require.Equal(t, "world", string(b[:n]))

	buf.Write([]byte("world"))
	ferr := errors.New("finish write error")
	buf.FinishWrite(ferr)

	b = make([]byte, 3)
	_, err = br.Read(b)
	require.NoError(t, err)
	require.Equal(t, "wor", string(b))

	b = make([]byte, 3)
	n, err = br.Read(b)
	require.Equal(t, ferr, err)
	require.Equal(t, "ld", string(b[:n]))

	buf.FinishWrite(nil)
	n, err = br.Read(b)
	require.Equal(t, io.EOF, err)
	require.Equal(t, "", string(b[:n]))

	isClosed := false
	br.closeFn = func() {
		isClosed = true
	}
	br.Close()
	require.True(t, isClosed)
}

func TestBufferReader_Copy(t *testing.T) {
	buf := &byteBuffer{}
	br := &bufferReader{buf: buf, readTimeout: 100 * time.Millisecond}

	b := make([]byte, 1<<20)
	io.ReadFull(rand.Reader, b)
	go func() {
		_, err := io.Copy(buf, bytes.NewReader(b))
		buf.FinishWrite(err)
	}()

	wbuf := bytes.NewBuffer(nil)
	io.Copy(wbuf, br)
	require.Equal(t, b, wbuf.Bytes())
}
