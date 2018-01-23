package sha1

import (
	"encoding/binary"
)

// Digest represents the partial evaluation of a checksum.
/*
type Digest struct {
	h   [5]uint32
	x   [_Chunk]byte
	nx  int
	len uint64
}
*/

func NewFromBytes(data *[96]byte) *Digest {

	d := new(Digest)

	d.h[0] = binary.LittleEndian.Uint32(data[0:])
	d.h[1] = binary.LittleEndian.Uint32(data[4:])
	d.h[2] = binary.LittleEndian.Uint32(data[8:])
	d.h[3] = binary.LittleEndian.Uint32(data[12:])
	d.h[4] = binary.LittleEndian.Uint32(data[16:])
	copy(d.x[:], data[20:])
	d.nx = int(binary.LittleEndian.Uint32(data[84:]))
	d.len = binary.LittleEndian.Uint64(data[88:])

	return d
}

func (d *Digest) ToBytes(data *[96]byte) {

	binary.LittleEndian.PutUint32(data[0:], d.h[0])
	binary.LittleEndian.PutUint32(data[4:], d.h[1])
	binary.LittleEndian.PutUint32(data[8:], d.h[2])
	binary.LittleEndian.PutUint32(data[12:], d.h[3])
	binary.LittleEndian.PutUint32(data[16:], d.h[4])
	copy(data[20:], d.x[:])
	binary.LittleEndian.PutUint32(data[84:], uint32(d.nx))
	binary.LittleEndian.PutUint64(data[88:], d.len)
}
