package log

import (
	"io"
	"os"
)

// ----------------------------------------------------------

type Logger struct {
	io.Writer
}

func New(f *os.File) Logger {
	return Logger{f}
}

func (r Logger) Log(msg []byte) {
	msg = append(msg, '\n')
	r.Write(msg)
}

// ----------------------------------------------------------

var Stdout = Logger{os.Stdout}
var Stderr = Logger{os.Stderr}

// ----------------------------------------------------------
