package mgotable

import (
	"errors"
	"reflect"
	"strings"
	"syscall"

	"github.com/qiniu/ds/table.config.v1"
	"gopkg.in/mgo.v2"
)

var (
	ErrInvalidSelector = errors.New("invalid selector")
	ErrInvalidIndex = errors.New("invalid index")
)

// ---------------------------------------------------------------------------

func ensureIndex(c *mgo.Collection, index string, unique bool) error {

	colIndexArr := strings.Split(index, ",")
	return c.EnsureIndex(mgo.Index{Key: colIndexArr, Unique: unique})
}

func ensureIndexByType(c *mgo.Collection, t reflect.Type, unique bool) error {

	n := t.NumField()
	fields := make([]string, n)
	for i := 0; i < n; i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("bson")
		if tag == "" {
			return ErrInvalidIndex
		}
		fields[i] = tag
	}

	return c.EnsureIndex(mgo.Index{Key: fields, Unique: unique})
}

// ---------------------------------------------------------------------------

func NewConfig(elem interface{}) *table.Config {

	return table.NewConfig(elem)
}

// ---------------------------------------------------------------------------

type Table struct {
	Coll *mgo.Collection
}

func New(coll *mgo.Collection, cfg *table.Config) (p Table, err error) {

	for _, t := range cfg.Uniques {
		err = ensureIndexByType(coll, t, true)
		if err != nil {
			return
		}
	}
	for _, t := range cfg.Indexes {
		err = ensureIndexByType(coll, t, false)
		if err != nil {
			return
		}
	}
	return Table{coll}, nil
}

func (p Table) RemoveAll(sel interface{}) (err error) {

	_, err = p.Coll.RemoveAll(sel)
	if err == mgo.ErrNotFound {
		return syscall.ENOENT
	}
	return
}

func (p Table) Insert(docs ...interface{}) (err error) {

	err = p.Coll.Insert(docs...)
	if mgo.IsDup(err) {
		return syscall.EEXIST
	}
	return
}

func (p Table) FindOne(ret interface{}, sel interface{}) (err error) {

	err = p.Coll.Find(sel).One(ret)
	if err == mgo.ErrNotFound {
		return syscall.ENOENT
	}
	return
}

func (p Table) FindAll(ret interface{}, sel interface{}) (err error) {

	err = p.Coll.Find(sel).All(ret)
	if err == mgo.ErrNotFound {
		return syscall.ENOENT
	}
	return
}

// ---------------------------------------------------------------------------

