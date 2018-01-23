package vm

// --------------------------------------------------------------------

type Range struct {
	Data []byte
	Addr uintptr
}

func (rg Range) Close() (err error) {
	return
}

// --------------------------------------------------------------------

func Map(fd int, offset int64, length int, prot int, flags int) (rg Range, err error) {

	panic("not impl")
}

// --------------------------------------------------------------------
