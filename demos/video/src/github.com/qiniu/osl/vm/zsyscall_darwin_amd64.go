package vm

import (
	. "syscall"
)

// --------------------------------------------------------------------

func mmap(addr uintptr, length uintptr, prot int, flag int, fd int, pos int64) (ret uintptr, err error) {
	r0, _, e1 := Syscall6(SYS_MMAP, uintptr(addr), uintptr(length), uintptr(prot), uintptr(flag), uintptr(fd), uintptr(pos))
	ret = uintptr(r0)
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
