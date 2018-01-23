package proto

// --------------------------------------------------------------------
// type Array

type Array interface {
	Puts(ielem int, elems []byte) (err error)
	Gets(ielem int, elems []byte) (n int, err error)
	Put(ielem int, elem []byte) (err error)
	Get(ielem int, elem []byte) (err error)
	Shrink(n int) (err error)
	Close() (err error)
	ElemLen() int
	Len() int
}

// --------------------------------------------------------------------
// type Bitmap

type Bitmap interface {
	Clear(idx int) error
	ClearRange(from, to int) error
	Set(idx int) error
	SetRange(from, to int) error
	Find(doClear bool) (idx int, err error)
	FindFrom(from int, doClear bool) (idx int, err error)
	Has(idx int) bool
	Close() (err error)
	DataOf() (data []uint64)
}

// --------------------------------------------------------------------
