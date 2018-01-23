package ino

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {

	fid, err := Stat("ino_test.go")
	if err != nil {
		t.Fatal("Stat failed:", err)
	}
	fmt.Println("fid:", fid)
}
