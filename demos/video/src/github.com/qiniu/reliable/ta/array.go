package ta

import (
	"encoding/binary"
	"github.com/qiniu/errors"
	rerr "github.com/qiniu/reliable/errors"
	"github.com/qiniu/reliable/proto"
	"io"
)

// --------------------------------------------------------------------
// type IArray

type IArray interface {
	proto.Array
	IUnderlayer
}

// --------------------------------------------------------------------
// type Array

type Array struct {
	IArray
	ta *Transaction
	id int
}

func OpenArray(arr IArray, ta *Transaction, id int) (p *Array, err error) {

	p = &Array{arr, ta, id}

	err = ta.init(id, p)
	if err != nil {
		err = errors.Info(err, "ta.OpenArray failed", id).Detail(err)
	}
	return
}

func (p *Array) Shrink(n int) (err error) {

	arr := p.IArray
	nn := arr.Len()
	if n >= nn {
		return nil
	}

	oldv := make([]byte, arr.ElemLen()*(nn-n))
	_, err = arr.Gets(n, oldv)
	if err != nil {
		if err != io.EOF {
			err = errors.Info(err, "ta.Array.Puts: array.Gets failed").Detail(err)
			return
		}
	}

	rl, hint, err := p.ta.beginRlog(p.id)
	if err != nil {
		err = errors.Info(err, "ta.Array.Put: beginRlog failed").Detail(err)
		return
	}
	rl.PutUint32(uint32(nn))
	rl.PutUint32(uint32(n))
	rl.Write(oldv)
	rl.end(hint)

	return arr.Shrink(n)
}

func (p *Array) Put(ielem int, elem []byte) (err error) {

	arr := p.IArray
	oldv := make([]byte, len(elem))
	nn := arr.Len()
	err = arr.Get(ielem, oldv)
	if err != nil {
		if err != io.EOF && !(err == rerr.ErrBadData && nn <= ielem) {
			err = errors.Info(err, "ta.Array.Put: array.Get failed").Detail(err)
			return
		}
	}

	rl, hint, err := p.ta.beginRlog(p.id)
	if err != nil {
		err = errors.Info(err, "ta.Array.Put: beginRlog failed").Detail(err)
		return
	}
	rl.PutUint32(uint32(nn))
	rl.PutUint32(uint32(ielem))
	rl.Write(oldv)
	rl.end(hint)

	return arr.Put(ielem, elem)
}

func (p *Array) Puts(ielem int, elems []byte) (err error) {

	arr := p.IArray
	oldv := make([]byte, len(elems))
	nn := arr.Len()
	n, err := arr.Gets(ielem, oldv)
	if err != nil {
		if err != io.EOF && !(err == rerr.ErrBadData && nn <= ielem) {
			err = errors.Info(err, "ta.Array.Puts: array.Gets failed").Detail(err)
			return
		}
	}

	rl, hint, err := p.ta.beginRlog(p.id)
	if err != nil {
		err = errors.Info(err, "ta.Array.Puts: beginRlog failed").Detail(err)
		return
	}
	rl.PutUint32(uint32(nn))
	rl.PutUint32(uint32(ielem))
	rl.Write(oldv[:n*arr.ElemLen()])
	rl.end(hint)

	err = arr.Puts(ielem, elems)
	return
}

func (p *Array) DoAct(act []byte) (err error) {

	arr := p.IArray
	nn := int(binary.LittleEndian.Uint32(act))
	ielem := int(binary.LittleEndian.Uint32(act[4:]))
	arr.Shrink(nn)
	return arr.Puts(ielem, act[8:])
}

// --------------------------------------------------------------------
