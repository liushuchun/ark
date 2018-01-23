package ioutil

import (
	"github.com/qiniu/log.v1"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/ts"
	"os"
	"testing"
)

// ---------------------------------------------------

func TestBasic(t *testing.T) {

	home := os.Getenv("HOME")
	files := []string{
		home + "/reliableTest1.txt",
		home + "/reliableTest2.txt",
		home + "/reliableTest3.txt",
	}
	for _, file := range files {
		os.Remove(file)
	}

	data, err := ReadFile(files)
	if err != ErrTooManyFails || data != nil {
		ts.Fatal(t, "ReadFile failed:", err)
	}

	err = WriteFile(files, []byte("Hello, world!\n"), 0666, 1)
	if err != nil {
		ts.Fatal(t, "WriteFile failed:", err)
	}

	data, err = ReadFile(files)
	if err != nil || string(data) != "Hello, world!\n" {
		ts.Fatal(t, "ReadFile failed:", string(data), err)
	}
}

// ---------------------------------------------------

func init() {
	log.SetOutputLevel(0)
}

// ---------------------------------------------------
