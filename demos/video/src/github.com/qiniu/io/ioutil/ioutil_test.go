package ioutil

import (
	"testing"
)

var (
	textFile = "text_file"
)

func TestReadFileLines(t *testing.T) {
	_, err := ReadFileLines(textFile, 10)
	if err == nil {
		t.Fatal("should return error if exceed")
	}

	lines, err := ReadFileLines(textFile, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 7 {
		t.Fatal("line num should be 7")
	}
}
