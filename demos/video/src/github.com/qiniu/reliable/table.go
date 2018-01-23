package reliable

import (
	"github.com/qiniu/encoding/binary"
	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	. "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/osl"
	"hash/crc32"
	"os"
)

/*
FileFormat:
	<fix-size-record><crc32>
	...
*/

// --------------------------------------------------------------------
// type Table

type Table struct {
	files      []osl.File
	rowlen     int
	allowfails int
}

func OpenTable(files []osl.File, rowlen, allowfails int) (p *Table, err error) {

	return &Table{files: files, rowlen: rowlen, allowfails: allowfails}, nil
}

func OpenTblfile(fnames []string, rowlen, allowfails int) (p *Table, err error) {

	files, err := osl.Open(fnames, allowfails)
	if err != nil {
		err = errors.Info(err, "OpenTblfile failed", fnames).Detail(err)
		return
	}
	p = &Table{files: files, rowlen: rowlen, allowfails: allowfails}
	return
}

func (p *Table) Close() (err error) {

	for _, f := range p.files {
		if f != nil {
			f.Close()
		}
	}
	p.files = nil
	return nil
}

func (p *Table) Underlayer() []osl.File {

	return p.files
}

func (p *Table) RowLen() int {

	return p.rowlen
}

func (p *Table) Rows() (rows int64, err error) {

	pos, err := osl.FsizeOf(p.files, p.allowfails)
	if err != nil {
		return
	}
	return pos / int64(p.rowlen+4), nil
}

func (p *Table) Stat() (fi os.FileInfo, err error) {

	pos, err := osl.FsizeOf(p.files, p.allowfails)
	if err != nil {
		return
	}
	return &osl.FileInfo{Fsize: pos}, nil
}

func (p *Table) Shrink(rows int64) (err error) {

	rows1, err := p.Rows()
	if err != nil {
		err = errors.Info(err, "reliable.Array.Shrink failed", rows).Detail(err)
		return
	}

	if rows >= rows1 {
		return nil
	}

	fsize := rows * int64(p.rowlen+4)
	for _, f := range p.files {
		if f != nil {
			err2 := f.Truncate(fsize)
			if err2 == nil {
				continue
			}
			log.Warn("reliable.Table.Shrink: Truncate failed -", err2)
			err = errors.Info(err2, "reliable.Table.Shrink: Truncate failed").Detail(err2)
			return
		}
	}
	return
}

func (p *Table) WriteRow(row int64, buf []byte) (err error) {

	n := p.rowlen
	if n != len(buf) {
		log.Warn("reliable.Table.WriteRow failed: invalid arguments")
		return ErrInvalidArgs
	}

	fails := 0
	allowfails := p.allowfails
	pos := row * int64(n+4)

	b := make([]byte, n+4)
	crc := crc32.ChecksumIEEE(buf)
	copy(b, buf)
	binary.LittleEndian.PutUint32(b[n:], crc)

	for _, f := range p.files {
		if f != nil {
			_, err2 := f.WriteAt(b, pos)
			if err2 == nil {
				continue
			}
			log.Warn("reliable.Table.WriteRow failed:", err2)
		}
		fails++
		if fails > allowfails {
			return ErrTooManyFails
		}
	}
	return
}

func (p *Table) WriteRows(row int64, bufs []byte) (err error) {

	n := p.rowlen
	if len(bufs)%n != 0 {
		log.Warn("reliable.Table.WriteRows failed: invalid arguments")
		return ErrInvalidArgs
	}
	rows := len(bufs) / n

	b := make([]byte, (n+4)*rows)
	roff := 0
	for off := 0; off < len(bufs); off += n {
		buf := bufs[off : off+n]
		crc := crc32.ChecksumIEEE(buf)
		copy(b[roff:], buf)
		binary.LittleEndian.PutUint32(b[roff+n:], crc)
		roff += n + 4
	}

	fails := 0
	allowfails := p.allowfails
	pos := row * int64(n+4)

	for _, f := range p.files {
		if f != nil {
			_, err2 := f.WriteAt(b, pos)
			if err2 == nil {
				continue
			}
			log.Warn("reliable.Table.WriteRows failed:", err2)
		}
		fails++
		if fails > allowfails {
			return ErrTooManyFails
		}
	}
	return
}

func (p *Table) ReadRow(row int64, buf []byte) (err error) {

	n := p.rowlen
	if n != len(buf) {
		log.Warn("reliable.Table.ReadRow failed: invalid arguments")
		return ErrInvalidArgs
	}

	b := make([]byte, n+4)
	pos := row * int64(n+4)

	return readRow(pos, b, buf, p.files, n)
}

func readRow(pos int64, b, buf []byte, files []osl.File, n int) (err error) {

	err = ErrBadData

	for i, f := range files {
		if f != nil {
			_, err2 := f.ReadAt(b, pos)
			if err2 != nil {
				log.Warn("reliable.Table.ReadRow failed:", err2)
				continue
			}
			crc := binary.LittleEndian.Uint32(b[n:])
			if crc32.ChecksumIEEE(b[:n]) != crc {
				if crc != 0 || notZeros(b[:n]) {
					log.Warn("reliable.Table.ReadRow failed: crc checksum error -", i, pos, crc)
					continue
				}
			}
			copy(buf, b)
			return nil
		}
	}
	return
}

func notZeros(b []byte) bool {

	for _, c := range b {
		if c != 0 {
			return true
		}
	}
	return false
}

func (p *Table) ReadRows(row int64, bufs []byte) (err error) {

	n := p.rowlen
	if len(bufs)%n != 0 {
		log.Warn("reliable.Table.ReadRows failed: invalid arguments")
		return ErrInvalidArgs
	}
	rows := len(bufs) / n

	bs := make([]byte, (n+4)*rows)
	pos := row * int64(n+4)

	for i, f := range p.files {
		if f != nil {
			_, err2 := f.ReadAt(bs, pos)
			if err2 != nil {
				log.Warn("reliable.Table.ReadRows failed:", err2)
				continue
			}
			roff := 0
			for off := 0; off < len(bufs); off += n {
				buf := bufs[off : off+n]
				b := bs[roff : roff+n+4]
				crc := binary.LittleEndian.Uint32(b[n:])
				if crc32.ChecksumIEEE(b[:n]) != crc {
					if crc != 0 || notZeros(b[:n]) {
						log.Info("ReadRows: crc checksum error -", i, pos+int64(roff), crc)
						err = readRow(pos+int64(roff), b, buf, p.files[i+1:], n)
						if err != nil {
							log.Error("reliable.Table.ReadRows failed:", err)
							return
						}
					}
				}
				copy(buf, b)
				roff += n + 4
			}
			return nil
		}
	}

	log.Error("reliable.Table.ReadRows failed: bad data")
	return ErrBadData
}

// --------------------------------------------------------------------
