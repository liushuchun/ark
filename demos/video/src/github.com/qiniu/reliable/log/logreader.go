package log

import (
	"bytes"
	"fmt"
	"github.com/qiniu/bufio"
	qbytes "github.com/qiniu/bytes"
	"github.com/qiniu/encoding/tab"
	qio "github.com/qiniu/io"
	"github.com/qiniu/log.v1"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/osl"
	"hash/crc32"
	"io"
	"strconv"
)

// --------------------------------------------------------------------
// func ReadFrom

func (p *Logger) ReadFrom(buf []byte, from int64) (n int, err error) {

	lastErr := io.EOF

	for i := 0; i < len(p.files); i++ {
		n1, err1 := p.files[i].ReadAt(buf, from)
		if err1 != nil && err1 != io.EOF {
			lastErr = err1
			continue
		}
		if n1 == 0 {
			return n, err1
		}
		n2, err2 := checkLines(buf[:n1])
		buf = buf[n2:]
		from += int64(n2)
		n += n2
		if err2 == nil {
			break
		}
		lastErr = err2
	}

	if n == 0 {
		err = lastErr
	}
	return
}

// --------------------------------------------------------------------
// type LogReader

type LogReader struct {
	r     *bufio.Reader
	sr    *io.SectionReader
	base  int64
	files []osl.File
}

func newLogReader(files []osl.File, pos, fsize int64, linemax int) *LogReader {

	for i, f := range files {
		if f == nil {
			continue
		}
		sr := io.NewSectionReader(f, pos, fsize-pos)
		r := bufio.NewReaderSize(sr, linemax)
		p := &LogReader{r: r, sr: sr, base: pos, files: files[i+1:]}
		return p
	}
	panic("not reachable")
}

func (p *LogReader) Tell() int64 {

	pos, _ := p.sr.Seek(0, 1)
	return pos + p.base - int64(p.r.Buffered())
}

func (p *LogReader) Scanln(a ...interface{}) (err error) {

	off := p.Tell()
	line, err := readLine(p.r)
	if err != nil {
		if err == io.EOF {
			return
		}
		line, err = p.repairReadLine(off)
		if err != nil {
			return
		}
	}

	r := qbytes.NewReader(line)
	_, err = fmt.Fscanln(r, a...)
	if err != nil {
		log.Error("reliable.LogReader.Readln: fmt.Fscanln failed -", err)
		return
	}

	err = tab.Unescapes(a)
	if err != nil {
		log.Error("reliable.LogReader.Readln: tab.Unescapes failed -", err)
	}
	return
}

func (p *LogReader) repairOffset(off int64) (err error) {

	delta := off - p.Tell()
	if delta < 0 {
		log.Error("repairOffset failed: bad offset")
		return ErrBadData
	}

	err = skip(p.r, int(delta))
	if err != nil {
		log.Error("repairOffset - skip failed:", err)
	}
	return
}

func (p *LogReader) repairReadLine(off int64) (line []byte, err error) {

	for i, f := range p.files {
		if f != nil {
			br := &qio.Reader{ReaderAt: f, Offset: off}
			r := bufio.NewReaderSize(br, p.r.BufferLen())
			line, err = readLine(r)
			if err == nil {
				off = br.Offset - int64(r.Buffered())
				err = p.repairOffset(off)
				return
			}
			log.Warn("repairReadLine failed:", i, err)
		}
	}

	log.Warn("repairReadLine failed: bad data")
	return nil, ErrBadData
}

func readLine(r *bufio.Reader) (line []byte, err error) {

	line, isPrefix, err := r.ReadLine()
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Warn("readLine failed:", err)
		return
	}
	if isPrefix {
		log.Warn("readLine failed: line too long")
		return nil, ErrLineTooLong
	}

	pos, err := checkLine(line)
	if err != nil {
		return
	}
	return line[pos+1:], nil
}

func checkLine(line []byte) (pos int, err error) {

	pos = bytes.IndexByte(line, '\t')
	if pos < 0 {
		log.Warn("checkLine crc failed: no seperator")
		return -1, ErrBadData
	}

	crc, err := strconv.ParseUint(string(line[:pos]), 36, 32)
	if err != nil {
		log.Warn("checkLine crc failed: invalid crc")
		return -1, ErrBadData
	}

	if crc32.ChecksumIEEE(line[pos+1:]) != uint32(crc) {
		log.Warn("checkLine failed: crc32 checksum error")
		return -1, ErrCrcChecksumError
	}
	return
}

func checkLines(lines []byte) (n int, err error) {

	for len(lines) > 0 {
		eol := bytes.IndexByte(lines, '\n')
		if eol < 0 {
			break
		}
		_, err = checkLine(lines[:eol])
		if err != nil {
			return
		}
		n += eol + 1
		lines = lines[eol+1:]
	}
	return
}

func skip(r *bufio.Reader, n int) (err error) {

	for {
		n1 := r.Buffered()
		if n <= n1 {
			_, err = r.Next(n)
			return
		}
		r.Next(n1)
		_, err = r.Peek(1)
		if err != nil {
			return
		}
		n -= n1
	}
}

// --------------------------------------------------------------------
