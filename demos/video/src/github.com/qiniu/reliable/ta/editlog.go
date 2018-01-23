package ta

import (
	"bytes"
	"encoding/binary"
	"github.com/qiniu/log.v1"
	"syscall"
)

// --------------------------------------------------------------------

type RevertLog struct {
	*bytes.Buffer
}

func (rl RevertLog) begin(id int) (rlog RevertLog, hint int, err error) {

	off := rl.Len()
	rl.WriteByte(byte(id))
	return rl, off, nil
}

func (rl RevertLog) PutUint32(v uint32) {

	var buf [4]byte
	b := buf[:]
	binary.LittleEndian.PutUint32(b, v)
	rl.Write(b)
}

func (rl RevertLog) end(hint int) {

	rl.PutUint32(uint32(hint))
}

func rollback(rl []byte, coms []IComponent) (err error) {

	n := len(rl)
	for n > 4 {
		off := int(binary.LittleEndian.Uint32(rl[n-4:]))
		if off >= n-4 {
			log.Warn("rollback: invalid rlog -", off, n-4)
			return syscall.EINVAL
		}
		id := int(rl[off])
		if id >= maxComponent {
			log.Warn("rollback: invalid rlog - id >= maxComponent", id)
			return syscall.EINVAL
		}
		com := coms[id]
		if com == nil {
			log.Warn("rollback: invalid rlog - nil component", id)
			return syscall.EINVAL
		}
		err = com.DoAct(rl[off+1 : n-4])
		if err != nil {
			log.Warn("rollback: com.DoAct failed -", err)
			return
		}
		n = off
	}
	if n != 0 {
		log.Warn("rollback: invalid rlog - off != 0", n)
		return syscall.EINVAL
	}
	return nil
}

// --------------------------------------------------------------------
