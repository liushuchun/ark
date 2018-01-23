package osl

import (
	"os"
)

func Permission(fi os.FileInfo) uint32 {
	return uint32(fi.Mode().Perm())
}

func Mtime(fi os.FileInfo) int64 {
	return fi.ModTime().UnixNano()
}

func IsRegular(fi os.FileInfo) bool {
	return (fi.Mode() & os.ModeType) == 0
}
