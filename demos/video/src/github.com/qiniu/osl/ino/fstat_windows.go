package ino

import (
	"encoding/base64"
	"os"
	"syscall"
	"unsafe"
)

func Fstat(fd uintptr) (fid string, err error) {

	var d syscall.ByHandleFileInformation
	err = syscall.GetFileInformationByHandle(syscall.Handle(fd), &d)
	if err == nil {
		var buf [12]byte
		*(*uint32)(unsafe.Pointer(&buf[0])) = d.FileIndexLow
		*(*uint32)(unsafe.Pointer(&buf[4])) = d.FileIndexHigh
		*(*uint32)(unsafe.Pointer(&buf[8])) = d.VolumeSerialNumber
		fid = base64.URLEncoding.EncodeToString(buf[:])
	}
	return
}

func FileIno(fname string, fi os.FileInfo) (fid string, err error) {

	return Stat(fname)
}
