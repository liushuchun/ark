package ioutil

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Random number state, accessed without lock; racy but harmless.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

// TempFile creates a new temporary file in the directory dir
// with a name beginning with prefix, opens the file for reading
// and writing, and returns the resulting *os.File.
// If dir is the empty string, TempFile uses the default directory
// for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously
// will not choose the same file.  The caller can use f.Name()
// to find the name of the file.  It is the caller's responsibility to
// remove the file when no longer needed.
func TempFile(dir, prefix, suffix string) (f *os.File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+nextSuffix()+suffix)
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				rand = reseed()
			}
			continue
		}
		break
	}
	return
}

func ReadLines(f io.Reader, lineLenMax int) (lines []string, err error) {
	var reader *bufio.Reader
	if lineLenMax > 0 {
		reader = bufio.NewReaderSize(f, lineLenMax)
	} else {
		reader = bufio.NewReader(f)
	}

	for {
		line, isPrefix, err := reader.ReadLine()
		if isPrefix {
			return nil, errors.New("Line length Exceeded")
		}
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil, err
		}

		lines = append(lines, string(line))
	}

	return
}

func ReadFileLines(fname string, lineLenMax int) (lines []string, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	return ReadLines(f, lineLenMax)
}
