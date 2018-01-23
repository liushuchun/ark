package table

import (
	"reflect"
)

// ---------------------------------------------------------------------------

type Config struct {
	ElemType reflect.Type
	Uniques  []reflect.Type
	Indexes  []reflect.Type
}

func NewConfig(elem interface{}) *Config {

	return &Config{
		ElemType: indirectType(elem),
	}
}

func (p *Config) WithUniques(uniques ...interface{}) *Config {

	p.Uniques = makeTypes(uniques)
	return p
}

func (p *Config) WithIndexes(indexes ...interface{}) *Config {

	p.Indexes = makeTypes(indexes)
	return p
}

func makeTypes(selectors []interface{}) []reflect.Type {

	types := make([]reflect.Type, len(selectors))
	for i, sel := range selectors {
		types[i] = indirectType(sel)
	}
	return types
}

func indirectType(val interface{}) reflect.Type {

	t := reflect.TypeOf(val)
	for {
		k := t.Kind()
		if k == reflect.Ptr {
			t = t.Elem()
			continue
		}
		if k != reflect.Struct {
			panic("data/selector must be struct")
		}
		return t
	}
}

// ---------------------------------------------------------------------------

