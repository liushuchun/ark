package vm

import (
	"syscall"
	"unsafe"
)

// --------------------------------------------------------------------

type slice struct {
	addr uintptr
	len  int
	cap  int
}

// --------------------------------------------------------------------

type Range []byte

func (rg Range) Close() (err error) {

	sl := (*slice)(unsafe.Pointer(&rg))
	return munmap(sl.addr, uintptr(sl.len))
}

// --------------------------------------------------------------------

func Map(fd int, offset int64, length int, prot int, flags int) (rg Range, err error) {

	if length <= 0 {
		return nil, syscall.EINVAL
	}

	// Map the requested memory.
	addr, errno := mmap(0, uintptr(length), prot, flags, fd, offset)
	if errno != nil {
		return nil, errno
	}

	// Slice memory layout
	var sl = slice{addr, length, length}

	// Use unsafe to turn sl into a []byte.
	b := *(*[]byte)(unsafe.Pointer(&sl))
	return b, nil
}

// --------------------------------------------------------------------
