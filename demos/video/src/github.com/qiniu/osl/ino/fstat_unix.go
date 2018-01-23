// +build !windows

package ino

import (
	"encoding/base64"
	"os"
	"syscall"
	"unsafe"
)

func Fstat(fd uintptr) (fid string, err error) {

	var stat syscall.Stat_t
	err = syscall.Fstat(int(fd), &stat)
	if err == nil {
		var buf [12]byte
		*(*uint64)(unsafe.Pointer(&buf[0])) = stat.Ino
		*(*uint32)(unsafe.Pointer(&buf[8])) = uint32(stat.Dev)
		fid = base64.URLEncoding.EncodeToString(buf[:])
	}
	return
}

func FileIno(fname string, fi os.FileInfo) (fid string, err error) {

	if sys := fi.Sys(); sys != nil {
		if stat, ok := sys.(syscall.Stat_t); ok {
			var buf [12]byte
			*(*uint64)(unsafe.Pointer(&buf[0])) = stat.Ino
			*(*uint32)(unsafe.Pointer(&buf[8])) = uint32(stat.Dev)
			fid = base64.URLEncoding.EncodeToString(buf[:])
			return
		}
	}
	return Stat(fname)
}
