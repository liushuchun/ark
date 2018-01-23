package table

import (
	"errors"
	"reflect"
	"syscall"

	"github.com/qiniu/ds/table.config.v1"
)

var (
	ErrInvalidSelector = errors.New("invalid selector")
	ErrInvalidDataType = errors.New("invalid data type")
)

var (
	zero reflect.Value
)

// ---------------------------------------------------------------------------

type selector struct {
	data      reflect.Value
	keyType   reflect.Type
	isUnique  bool
	isRemoved bool
}

func newSelector(keyType, elemType reflect.Type, isUnique bool) *selector {

	var mapType reflect.Type

	elemPtrType := reflect.PtrTo(elemType)
	if isUnique {
		mapType = reflect.MapOf(keyType, elemPtrType)
	} else {
		mapType = reflect.MapOf(keyType, reflect.SliceOf(elemPtrType))
	}
	return &selector{
		data: reflect.MakeMap(mapType),
		keyType: keyType,
		isUnique: isUnique,
	}
}

func (p *selector) keyOf(doc reflect.Value) reflect.Value {

	t := p.keyType
	v := reflect.New(t).Elem()
	n := t.NumField()
	for i := 0; i < n; i++ {
		sf := t.Field(i)
		sv := doc.FieldByName(sf.Name)
		v.Field(i).Set(sv)
	}
	return v
}

func (p *selector) Insert(doc reflect.Value) (err error) {

	data := p.data
	key := p.keyOf(doc)
	if p.isUnique {
		v := data.MapIndex(key)
		if v.IsValid() {
			return syscall.EEXIST
		}
		p.data.SetMapIndex(key, doc.Addr())
	} else {
		old := data.MapIndex(key)
		if !old.IsValid() {
			elemPtrType := reflect.PtrTo(doc.Type())
			sliceType := reflect.SliceOf(elemPtrType)
			old = reflect.MakeSlice(sliceType, 0, 1)
		}
		old = reflect.Append(old, doc.Addr())
		data.SetMapIndex(key, old)
	}
	return
}

func (p *selector) RemoveDoc(pdoc reflect.Value) {

	data := p.data
	key := p.keyOf(pdoc.Elem())
	if p.isUnique {
		data.SetMapIndex(key, zero)
	} else {
		all := data.MapIndex(key)
		m := all.Len()
		docPtr := pdoc.Pointer()
		for j := 0; j < m; j++ {
			if all.Index(j).Pointer() == docPtr {
				all = reflect.AppendSlice(all.Slice(0, j), all.Slice(j+1, m))
				data.SetMapIndex(key, all)
				return
			}
		}
	}
}

func (p *selector) Find(sel reflect.Value) (ret reflect.Value, err error) {

	ret = p.data.MapIndex(sel)
	if !ret.IsValid() {
		err = syscall.ENOENT
	}
	return
}

func (p *selector) Remove(sel reflect.Value) (ret reflect.Value, err error) {

	data := p.data
	ret = data.MapIndex(sel)
	data.SetMapIndex(sel, zero)
	if !ret.IsValid() {
		err = syscall.ENOENT
	}
	return
}

// ---------------------------------------------------------------------------

func NewConfig(elem interface{}) *table.Config {

	return table.NewConfig(elem)
}

// ---------------------------------------------------------------------------

type Table struct {
	sels     map[reflect.Type]*selector
	elemType reflect.Type
}

func New(cfg *table.Config) *Table {

	elemType := cfg.ElemType
	sels := make(map[reflect.Type]*selector)
	for _, unique := range cfg.Uniques {
		psel := newSelector(unique, elemType, true)
		sels[unique] = psel
	}
	for _, index := range cfg.Indexes {
		psel := newSelector(index, elemType, false)
		sels[index] = psel
	}
	return &Table{
		elemType: cfg.ElemType,
		sels: sels,
	}
}

func (p *Table) Insert(docs ...interface{}) (err error) {

	for _, doc := range docs {
		v := reflect.Indirect(reflect.ValueOf(doc))
		if v.Type() != p.elemType {
			return ErrInvalidDataType
		}
		if !v.CanAddr() {
			t := reflect.New(p.elemType).Elem()
			t.Set(v)
			v = t
		}
		for _, sel := range p.sels {
			err = sel.Insert(v)
			if err != nil {
				return
			}
		}
	}
	return
}

func (p *Table) FindOne(ret interface{}, sel interface{}) (err error) {

	vsel := reflect.Indirect(reflect.ValueOf(sel))
	psel, ok := p.sels[vsel.Type()]
	if !ok || !psel.isUnique {
		return ErrInvalidSelector
	}

	retRef := reflect.ValueOf(ret)
	if retRef.Kind() != reflect.Ptr {
		return ErrInvalidDataType
	}
	retRef = retRef.Elem()

	vret, err := psel.Find(vsel)
	if err != nil {
		return
	}

	elemPtrType := reflect.PtrTo(p.elemType)
	switch retRef.Type() {
	case elemPtrType: // *Type
		retRef.Set(vret)
	case p.elemType:  // Type
		retRef.Set(vret.Elem())
	default:
		return ErrInvalidDataType
	}
	return
}

func (p *Table) FindAll(ret interface{}, sel interface{}) (err error) {

	vsel := reflect.Indirect(reflect.ValueOf(sel))
	psel, ok := p.sels[vsel.Type()]
	if !ok || psel.isUnique {
		return ErrInvalidSelector
	}

	retRef := reflect.ValueOf(ret)
	if retRef.Kind() != reflect.Ptr {
		return ErrInvalidDataType
	}
	retRef = retRef.Elem()

	vret, err := psel.Find(vsel)
	if err != nil {
		return
	}

	elemPtrType := reflect.PtrTo(p.elemType)
	switch retRef.Type() {
	case reflect.SliceOf(elemPtrType): // []*Type
		retRef.Set(vret)
	case reflect.SliceOf(p.elemType):  // []Type
		n := vret.Len()
		sliceType := reflect.SliceOf(p.elemType)
		arr := reflect.MakeSlice(sliceType, n, n)
		for i := 0; i < n; i++ {
			arr.Index(i).Set(vret.Index(i).Elem())
		}
		retRef.Set(arr)
	default:
		return ErrInvalidDataType
	}
	return
}

func (p *Table) RemoveAll(sel interface{}) (err error) {

	vsel := reflect.Indirect(reflect.ValueOf(sel))
	psel, ok := p.sels[vsel.Type()]
	if !ok {
		return ErrInvalidSelector
	}

	vret, err := psel.Remove(vsel)
	if err != nil {
		return
	}

	psel.isRemoved = true
	for _, t := range p.sels {
		if !t.isRemoved {
			if psel.isUnique {
				t.RemoveDoc(vret)
			} else {
				n := vret.Len()
				for i := 0; i < n; i++ {
					t.RemoveDoc(vret.Index(i))
				}
			}
		}
	}
	psel.isRemoved = false
	return
}

// ---------------------------------------------------------------------------

