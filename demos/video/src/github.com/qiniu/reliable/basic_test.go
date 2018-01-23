package reliable

import (
	"github.com/qiniu/log.v1"
	rts "github.com/qiniu/reliable/ts"
)

// ---------------------------------------------------

const (
	writeOk    = rts.WriteOk
	writeFail  = rts.WriteFail
	writeBad   = rts.WriteBad
	writeShort = rts.WriteShort
)

func newBuffer(modes []int) *rts.Buffer {

	return rts.NewBuffer(modes)
}

// ---------------------------------------------------

func makeSlice(s string, rowlen int) []byte {

	b := make([]byte, rowlen)
	copy(b, s)
	return b
}

func makeSlices(ss []string, rowlen int) []byte {

	b := make([]byte, rowlen*len(ss))
	off := 0
	for _, s := range ss {
		copy(b[off:off+rowlen], s)
		off += rowlen
	}
	return b
}

// ---------------------------------------------------

func init() {
	log.SetOutputLevel(0)
}

// ---------------------------------------------------
