package binary

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

type Foo struct {
	A    uint8    // 1 Byte
	B    uint8    // 1 Byte (LargeKey+ChunkBits)
	C    uint16   // 2 Byte
	Keys []uint32 // n * 4 Byte
	Vals []uint64 // n * 8 Byte
}

func TestExt(t *testing.T) {

	foo := &Foo{
		C:    32,
		Keys: []uint32{1, 2, 3},
		Vals: []uint64{5, 6, 7, 8},
	}

	b := bytes.NewBuffer(nil)
	err := Write(b, LittleEndian, foo)
	if err != nil {
		t.Fatal("Write failed:", err)
	}

	foo2 := &Foo{
		Keys: make([]uint32, 3),
		Vals: make([]uint64, 4),
	}
	err = Read(b, LittleEndian, foo2)
	if err != nil || !reflect.DeepEqual(foo, foo2) {
		t.Fatal("Read failed:", foo2, err)
	}
}

func TestOldExt(t *testing.T) {

	foo := &Foo{
		C:    32,
		Keys: []uint32{1, 2, 3},
		Vals: []uint64{5, 6, 7, 8},
	}

	b := bytes.NewBuffer(nil)
	err := binary.Write(b, LittleEndian, foo)
	if err == nil {
		t.Fatal("binary.Write ok?")
	}
}
