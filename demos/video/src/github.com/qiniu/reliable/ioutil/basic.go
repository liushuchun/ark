package ioutil

import (
	"github.com/qiniu/log.v1"
	. "github.com/qiniu/reliable/errors"
	"hash/crc32"
	"io/ioutil"
	"os"
)

/*
FileFormat:
	<crc32-hex-8-bytes-string>\r\n
	...
*/

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

func loadFile(fname string) ([]byte, error) {

	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Warn("reliable.loadFile failed:", fname, err)
		return nil, err
	}

	crc, err := loadCrc(data)
	if err != nil {
		return nil, err
	}

	if crc32.ChecksumIEEE(data[10:]) != crc {
		log.Warn("reliable.loadFile failed: crc checksum error")
		return nil, ErrCrcChecksumError
	}

	return data[10:], nil
}

func ReadFile(fnames []string) ([]byte, error) {

	for _, fname := range fnames {
		data, err := loadFile(fname)
		if err == nil {
			return data, nil
		}
	}
	return nil, ErrTooManyFails
}

// ---------------------------------------------------

var hexchars = "0123456789abcdef"

func WriteFile(fnames []string, data []byte, perm os.FileMode, allowfails int) error {

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
	for i, fname := range fnames {
		err := ioutil.WriteFile(fname, b, perm)
		if err != nil {
			log.Warn("reliable.WriteFile failed:", i, fname, err)
			fails++
			if fails > allowfails {
				return ErrTooManyFails
			}
		}
	}
	return nil
}

// ---------------------------------------------------
