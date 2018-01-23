package reliable

import (
	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/osl"
	"hash/crc32"
	"syscall"
)

/*
FileFormat:
	<crc32-hex-8-bytes-string>\r\n
	...
*/

// ---------------------------------------------------

func ReadAll(f osl.File) (data []byte, err error) {

	fi, err := f.Stat()
	if err != nil {
		err = errors.Info(err, "ReadAll: f.Stat failed").Detail(err)
		return
	}

	fsize := fi.Size()
	if (fsize >> 32) != 0 {
		return nil, syscall.ENOMEM
	}
	n := int(fsize)

	data = make([]byte, n)
	_, err = f.ReadAt(data, 0)
	if err != nil {
		err = errors.Info(err, "ReadAll: f.ReadAt failed").Detail(err)
	}
	return
}

func WriteAll(f osl.File, data []byte) (err error) {

	_, err = f.WriteAt(data, 0)
	if err != nil {
		err = errors.Info(err, "WriteAll: f.WriteAt failed").Detail(err)
		return
	}

	err = f.Truncate(int64(len(data)))
	if err != nil {
		err = errors.Info(err, "WriteAll: f.Truncate failed").Detail(err)
	}
	return
}

// ---------------------------------------------------

func loadCrc(data []byte) (crc uint32, err error) {

	if len(data) < 10 || data[8] != '\r' || data[9] != '\n' {
		log.Warn("reliable.loadCrc failed: no crc32 header")
		return 0, ErrBadData
	}

	for i := 0; i < 8; i++ {
		c := data[i]
		if c >= '0' && c <= '9' {
			crc = (crc << 4) | uint32(c-'0')
		} else if c >= 'a' && c <= 'f' {
			crc = (crc << 4) | uint32(c-('a'-10))
		} else {
			log.Warn("reliable.loadCrc failed: invalid crc32 hex string")
			return 0, ErrBadData
		}
	}
	return
}

func loadFile(f osl.File) ([]byte, error) {

	data, err := ReadAll(f)
	if err != nil {
		err = errors.Info(err, "reliable.loadFile: ReadAll failed").Detail(err)
		return nil, err
	}

	crc, err := loadCrc(data)
	if err != nil {
		err = errors.Info(err, "reliable.loadFile: loadCrc failed").Detail(err)
		return nil, err
	}

	if crc32.ChecksumIEEE(data[10:]) != crc {
		log.Warn("reliable.loadFile failed: crc checksum error")
		return nil, ErrCrcChecksumError
	}

	return data[10:], nil
}

// ---------------------------------------------------
// type Config

type Config struct {
	files      []osl.File
	allowfails int
}

func OpenConfig(files []osl.File, allowfails int) (p *Config, err error) {

	return &Config{files: files, allowfails: allowfails}, nil
}

func OpenCfgfile(fnames []string, allowfails int) (p *Config, err error) {

	files, err := osl.Open(fnames, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenCfgfile failed", fnames).Detail(err)
		return
	}

	return OpenConfig(files, allowfails)
}

func (p *Config) Underlayer() []osl.File {

	return p.files
}

func (p *Config) Close() (err error) {

	for _, f := range p.files {
		if f != nil {
			f.Close()
		}
	}
	p.files = nil
	return nil
}

func (p *Config) Validate() (err error) {

	fsize, err := osl.FsizeOf(p.files, p.allowfails)
	if err != nil {
		return
	}
	if fsize == 0 {
		return syscall.ENOENT
	}
	return nil
}

func (p *Config) ReadFile() ([]byte, error) {

	for _, f := range p.files {
		data, err := loadFile(f)
		if err == nil {
			return data, nil
		}
	}
	return nil, ErrTooManyFails
}

// ---------------------------------------------------

var hexchars = "0123456789abcdef"

func (p *Config) WriteFile(data []byte) error {

	b := make([]byte, len(data)+10)

	crc := crc32.ChecksumIEEE(data)
	for i := 7; i >= 0; i-- {
		b[i] = hexchars[crc&0x0f]
		crc >>= 4
	}
	b[8] = '\r'
	b[9] = '\n'
	copy(b[10:], data)

	fails := 0
	for i, f := range p.files {
		err := WriteAll(f, b)
		if err != nil {
			log.Warn("reliable.WriteFile failed:", i, err)
			fails++
			if fails > p.allowfails {
				return ErrTooManyFails
			}
		}
	}
	return nil
}

// ---------------------------------------------------
