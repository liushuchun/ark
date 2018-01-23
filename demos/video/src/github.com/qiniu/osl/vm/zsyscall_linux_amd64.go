package vm

import (
	. "syscall"
)

// --------------------------------------------------------------------

func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (xaddr uintptr, err error) {
	r0, _, e1 := Syscall6(SYS_MMAP, uintptr(addr), uintptr(length), uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset))
	xaddr = uintptr(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func munmap(addr uintptr, length uintptr) (err error) {
	_, _, e1 := Syscall(SYS_MUNMAP, uintptr(addr), uintptr(length), 0)
	if e1 != 0 {
		err = e1
	}
	return
}

// --------------------------------------------------------------------
