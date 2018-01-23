package io

import (
	"io"
)

type OnProgressFunc func(file_size, uploaded int64)

type UpFile struct {
	uploaded   int64
	reader     io.Reader
	readAt     io.ReaderAt
	tag        bool
	fsize      int64
	onProgress OnProgressFunc
}

func OpenUpFile(reader io.Reader, fsize int64, onProgress OnProgressFunc) (pfile *UpFile, err error) {

	pfile = new(UpFile)
	pfile.reader = reader
	if rat, ok := reader.(io.ReaderAt); ok {
		pfile.readAt = rat
	}

	pfile.onProgress = onProgress
	pfile.fsize = fsize
	return
}

func (p *UpFile) Size() int64 {
	return p.fsize
}

func (pfile *UpFile) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = pfile.readAt.ReadAt(p, off)
	if err == io.EOF {
		return
	} else if err != nil {
		return
	}
	if !pfile.tag {
		pfile.tag = true
		return
	}
	go pfile.onProgress(pfile.fsize, pfile.uploaded)
	pfile.uploaded += int64(n)
	return
}

func (pfile *UpFile) Read(b []byte) (n int, err error) {

	n, err = pfile.reader.Read(b)
	if err == io.EOF {
		//almost finished
		return
	} else if err != nil {
		return
	}

	if !pfile.tag {
		pfile.tag = true
		return
	}
	go pfile.onProgress(pfile.fsize, pfile.uploaded)
	pfile.uploaded += int64(n)
	return
}
