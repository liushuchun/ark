package ioutil

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestSeqWriter1(t *testing.T) {

	b := bytes.NewBuffer(nil)
	sw := SeqWriter(ioutil.Discard, 3, b, 4)
	text := "Hello, golang world!"

	n, err := io.Copy(sw, strings.NewReader(text))
	println(string(b.Bytes()))

	if err != nil || n != int64(len(text)) {
		t.Fatal("io.Copy failed:", n, err)
	}
	if b.Len() != 4 || string(b.Bytes()) != text[3:7] {
		t.Fatal("b:", string(b.Bytes()), b.Len())
	}
}

func TestSeqWriter2(t *testing.T) {

	b := bytes.NewBuffer(nil)
	sw := SeqWriter(ioutil.Discard, 3, b, 4)
	text := "Hello, golang world!"

	sw.Write([]byte(text[:2]))
	sw.Write([]byte(text[2:4]))
	sw.Write([]byte(text[4:]))
	println(string(b.Bytes()))

	if b.Len() != 4 || string(b.Bytes()) != text[3:7] {
		t.Fatal("b:", string(b.Bytes()), b.Len())
	}
}
