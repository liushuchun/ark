package tableutil

import (
	"github.com/qiniu/bytes"
	"github.com/qiniu/encoding/binary"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/reliable"
	"io"
)

func WriteRow(p *reliable.Table, row int64, v interface{}) (err error) {

	rowlen := p.RowLen()
	buf := make([]byte, rowlen)

	w := bytes.NewWriter(buf)
	err = binary.Write(w, binary.LittleEndian, v)
	if err != nil {
		log.Error("WriteRow failed:", err, v)
		return
	}
	if w.Len() != rowlen {
		log.Error("WriteRow failed: short write -", v)
		return io.ErrShortWrite
	}

	return p.WriteRow(row, buf)
}

func WriteRows(p *reliable.Table, row int64, vals []interface{}) (err error) {

	rowlen := p.RowLen()
	buf := make([]byte, rowlen*len(vals))
	w := bytes.NewWriter(buf)

	rowslen := 0
	for _, v := range vals {
		err = binary.Write(w, binary.LittleEndian, v)
		if err != nil {
			log.Error("WriteRow failed:", err, v)
			return
		}
		rowslen += rowlen
		if w.Len() != rowslen {
			log.Error("WriteRow failed: short write -", v)
			return io.ErrShortWrite
		}
	}

	return p.WriteRows(row, buf)
}

func ReadRow(p *reliable.Table, row int64, v interface{}) (err error) {

	rowlen := p.RowLen()
	buf := make([]byte, rowlen)

	err = p.ReadRow(row, buf)
	if err != nil {
		return
	}

	r := bytes.NewReader(buf)
	err = binary.Read(r, binary.LittleEndian, v)
	if err != nil {
		log.Error("ReadRow failed:", err)
		return
	}
	if r.Len() != 0 {
		log.Error("ReadRow failed: unexpected eof")
		return io.ErrUnexpectedEOF
	}
	return
}

func ReadRows(p *reliable.Table, row int64, vals []interface{}) (err error) {

	rowlen := p.RowLen()
	rowslen := rowlen * len(vals)
	buf := make([]byte, rowslen)

	err = p.ReadRows(row, buf)
	if err != nil {
		return
	}

	r := bytes.NewReader(buf)

	for _, v := range vals {
		err = binary.Read(r, binary.LittleEndian, v)
		if err != nil {
			log.Error("ReadRow failed:", err)
			return
		}
		rowslen -= rowlen
		if r.Len() != rowslen {
			log.Error("ReadRow failed: unexpected eof")
			return io.ErrUnexpectedEOF
		}
	}
	return
}

// --------------------------------------------------------------------
