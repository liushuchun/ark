package ino

import (
	"os"
)

func Stat(name string) (fid string, err error) {

	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	return Fstat(f.Fd())
}
