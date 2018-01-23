package log

import (
	"bytes"
	"fmt"
	"github.com/qiniu/encoding/tab"
	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/osl"
	"hash/crc32"
	"os"
	"strconv"
	"sync"
)

/*
FileFormat:
	<crc32(not-include-eol)-in-base36>\t<log-message-in-one-row>\n
	...
*/

const minLineMax = 4096

// --------------------------------------------------------------------
// type Logger

type Logger struct {
	files               []osl.File
	pos                 int64
	mutex               sync.Mutex
	allowfails, linemax int
}

func Open(fnames []string, linemax, allowfails int) (p *Logger, err error) {

	files, err := osl.Open(fnames, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenLogfile failed", fnames).Detail(err)
		return
	}

	p = &Logger{files: files, allowfails: allowfails, linemax: linemax}
	return p.init(allowfails)
}

func OpenEx(files []osl.File, linemax, allowfails int) (p *Logger, err error) {

	p = &Logger{files: files, allowfails: allowfails, linemax: linemax}
	return p.init(allowfails)
}

func (p *Logger) Underlayer() []osl.File {

	return p.files
}

func (p *Logger) Close() (err error) {

	for _, f := range p.files {
		if f != nil {
			f.Close()
		}
	}
	p.files = nil
	return nil
}

func (p *Logger) Stat() (fi os.FileInfo, err error) {

	p.mutex.Lock()
	pos := p.pos
	p.mutex.Unlock()

	return &osl.FileInfo{Fsize: pos}, nil
}

func (p *Logger) init(allowfails int) (*Logger, error) {

	pos := int64(0)
	fails := 0

	if p.linemax < minLineMax {
		p.linemax = minLineMax
	}

	buf := make([]byte, p.linemax)
	for _, f := range p.files {
		if f != nil {
			pos2, err2 := logFileCheck(f, buf)
			if err2 == nil {
				if pos < pos2 {
					pos = pos2
				}
				continue
			}
			log.Warn("reliable.Logger.init failed:", err2)
		}
		fails++
		if fails > allowfails {
			p.Close()
			return nil, ErrTooManyFails
		}
	}

	p.pos = pos
	return p, nil
}

var logSep = []byte{'\n'}

func logFileCheck(f osl.File, buf []byte) (fsize int64, err error) {

	fi, err := f.Stat()
	if err != nil {
		log.Warn("reliable.logFileCheck: file.Stat failed -", err)
		return
	}

	fsize = fi.Size()
	if fsize == 0 {
		return
	}

	from := fsize - int64(len(buf))
	if from < 0 {
		from = 0
		buf = buf[:int(fsize)]
	}

	n, err := f.ReadAt(buf, from)
	if err != nil && n != len(buf) {
		log.Warn("reliable.logFileCheck: file.ReadAt failed -", err)
		return
	}

	if buf[n-1] == '\n' {
		return
	}

	pos := bytes.LastIndex(buf, logSep)
	if pos < 0 {
		log.Warn("reliable.logFileCheck failed: line too long")
		err = ErrLineTooLong
		return
	}

	log.Info("logFileCheck repaired: truncate -", buf[pos+1:])

	return from + int64(pos+1), nil
}

func (p *Logger) Reader(from int64) *LogReader {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return newLogReader(p.files, from, p.pos, p.linemax)
}

func (p *Logger) write(line []byte) (err error) {

	fails := 0
	nl := len(line)
	crc := crc32.ChecksumIEEE(line[:nl-1])
	crcstr := strconv.FormatUint(uint64(crc), 36)
	ncrc := len(crcstr)

	nlen := ncrc + 1 + nl
	if nlen > p.linemax {
		return ErrLineTooLong
	}

	buf := make([]byte, nlen)
	copy(buf, crcstr)
	buf[ncrc] = '\t'
	copy(buf[ncrc+1:], line)

	allowfails := p.allowfails

	p.mutex.Lock()
	defer p.mutex.Unlock()

	pos := p.pos
	for _, f := range p.files {
		if f != nil {
			_, err2 := f.WriteAt(buf, pos)
			if err2 == nil {
				continue
			}
			log.Warn("reliable.Writeln failed:", err2)
		}
		fails++
		if fails > allowfails {
			return ErrTooManyFails
		}
	}
	p.pos = pos + int64(nlen)
	return
}

func (p *Logger) Log(msg []byte) (err error) {

	msg = append(msg, '\n')
	return p.write(msg)
}

func (p *Logger) Println(a ...interface{}) (err error) {

	a = tab.Escapes(a)
	line := fmt.Sprintln(a...)
	return p.write([]byte(line))
}

// --------------------------------------------------------------------
