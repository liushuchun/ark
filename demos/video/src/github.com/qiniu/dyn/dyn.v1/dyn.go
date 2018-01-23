package dyn

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/qiniu/log.v1"
)

// ----------------------------------------------------------

func TagName(tag string) string {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}

func FieldByTag(cate string, sv reflect.Value, tagName string) (v reflect.Value) {

	st := sv.Type()
	for i := 0; i < sv.NumField(); i++ {
		sf := st.Field(i)
		tag := sf.Tag.Get(cate)
		if TagName(tag) == tagName {
			v = sv.Field(i)
			return
		}
	}
	return
}

// ----------------------------------------------------------

func getCmdVal(data reflect.Value, cmd string) (v reflect.Value, ok bool) {

	execCmd := data.MethodByName("ExecCmd") // v, err := this.ExecCmd(cmd)
	if !execCmd.IsValid() {
		if data.Kind() == reflect.Struct {
			log.Warn("GetVal: method ExecCmd not foud -", reflect.TypeOf(data.Interface()))
		}
		return
	}

	in := []reflect.Value{reflect.ValueOf(cmd)}
	out := execCmd.Call(in)
	if len(out) > 1 {
		if !out[1].IsNil() {
			log.Warn("GetVal: ExecCmd failed -", cmd, out[1].Interface())
			return
		}
	}
	v, ok = out[0], true
	return
}

func GetVal(cate string, data reflect.Value, key string) (v reflect.Value, ok bool) {

	parts := strings.Split(key, ".")

	for _, part := range parts {
		if strings.HasPrefix(part, "`") {
			n := len(part) - 1
			if n < 1 || part[n] != '`' {
				log.Warn("GetVal: invalid part -", part)
				return
			}
			ok1 := false
			if data.Kind() == reflect.Interface {
				data = data.Elem()
			}
			if v, ok1 = getCmdVal(data, part[1:n]); !ok1 {
				return
			}
			data = v
			continue
		}
	retry:
		kind := data.Kind()
		switch kind {
		case reflect.Ptr, reflect.Interface:
			data = data.Elem()
			goto retry
		case reflect.Struct:
			v = FieldByTag(cate, data, part)
		case reflect.Map:
			v = data.MapIndex(reflect.ValueOf(part))
		case reflect.Array, reflect.Slice:
			index, err := strconv.Atoi(part)
			if err != nil {
				log.Warn("GetVal failed: unsupported index -", part)
				return
			}
			v = data.Index(index)
		case reflect.Func:
			out := data.Call(nil)
			if len(out) != 1 {
				log.Warn("GetVal failed: unsupport type -", data.Type(), ", key:", part)
				return
			}
			data = out[0]
			log.Debug("deref:", data.Interface(), part)
			goto retry
		case reflect.Invalid:
			return
		default:
			log.Warn("GetVal failed: unsupported type -", kind, data.Type(), ", key:", part)
			return
		}
		if !v.IsValid() {
			ok1 := false
			if v, ok1 = getCmdVal(data, part); !ok1 {
				log.Warn("GetVal failed: not found key -", part)
				return
			}
		}
		data = v
	}
	ok = true
	return
}

// ----------------------------------------------------------

func Get(data interface{}, key string) (v interface{}, ok bool) {

	val, ok := GetVal("json", reflect.ValueOf(data), key)
	if ok {
		v = val.Interface()
	}
	return
}

// ----------------------------------------------------------

func Int(data interface{}) (val int64, ok bool) {

retry:
	switch v := data.(type) {
	case int:
		val = int64(v)
	case uint:
		val = int64(v)
	case int64:
		val = v
	case uint64:
		val = int64(v)
	case uintptr:
		val = int64(v)
	case int32:
		val = int64(v)
	case uint32:
		val = int64(v)
	case int16:
		val = int64(v)
	case uint16:
		val = int64(v)
	case uint8:
		val = int64(v)
	case int8:
		val = int64(v)
	case func() interface{}:
		data = v()
		goto retry
	default:
		return
	}
	ok = true
	return
}

func GetInt(data interface{}, key string) (val int64, ok bool) {

	v, ok := Get(data, key)
	if ok {
		val, ok = Int(v)
	}
	return
}

// ----------------------------------------------------------

func Float(data interface{}) (val float64, ok bool) {

retry:
	switch v := data.(type) {
	case float64:
		val = v
	case float32:
		val = float64(v)
	case func() interface{}:
		data = v()
		goto retry
	default:
		v2, ok2 := Int(data)
		if !ok2 {
			return
		}
		val = float64(v2)
	}
	ok = true
	return
}

func GetFloat(data interface{}, key string) (val float64, ok bool) {

	v, ok := Get(data, key)
	if ok {
		val, ok = Float(v)
	}
	return
}

// ----------------------------------------------------------

func String(data interface{}) (val string, ok bool) {

retry:
	switch v := data.(type) {
	case string:
		val = v
	case func() interface{}:
		data = v()
		goto retry
	default:
		return
	}
	ok = true
	return
}

func GetString(data interface{}, key string) (val string, ok bool) {

	v, ok := Get(data, key)
	if ok {
		val, ok = String(v)
	}
	return
}

// ----------------------------------------------------------
