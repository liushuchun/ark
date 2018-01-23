package osl

import (
	"os"
	"time"
)

func Chtimes(name string, atime int64, mtime int64) error {
	atime1 := time.Unix(0, atime)
	mtime1 := time.Unix(0, mtime)
	return os.Chtimes(name, atime1, mtime1)
}
