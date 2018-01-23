package object

import (
	"strconv"
)

// Value 帮助你从任意 interface{} 转换到各种值类型
// Must* 系列返回的值，源类型复合的情况下才返回
// 其他返回错误的系列，是强转换
type Value struct {
	value interface{}
}

func (v *Value) Bool() (bool, error) {
	if ret, ok := v.value.(bool); ok {
		return ret, nil
	}
	if v.value == "on" {
		return true, nil
	}
	return strconv.ParseBool(v.String())
}

func (v *Value) Float32() (float32, error) {
	if ret, ok := v.value.(float32); ok {
		return ret, nil
	}
	value, err := strconv.ParseFloat(v.String(), 32)
	return float32(value), err
}

func (v *Value) Float64() (float64, error) {
	if ret, ok := v.value.(float64); ok {
		return ret, nil
	}
	return strconv.ParseFloat(v.String(), 64)
}

func (v *Value) Int() (int, error) {
	if ret, ok := v.value.(int); ok {
		return ret, nil
	}
	value, err := strconv.ParseInt(v.String(), 10, 32)
	return int(value), err
}

func (v *Value) Int8() (int8, error) {
	if ret, ok := v.value.(int8); ok {
		return ret, nil
	}
	value, err := strconv.ParseInt(v.String(), 10, 8)
	return int8(value), err
}

func (v *Value) Int16() (int16, error) {
	if ret, ok := v.value.(int16); ok {
		return ret, nil
	}
	value, err := strconv.ParseInt(v.String(), 10, 16)
	return int16(value), err
}

func (v *Value) Int32() (int32, error) {
	if ret, ok := v.value.(int32); ok {
		return ret, nil
	}
	value, err := strconv.ParseInt(v.String(), 10, 32)
	return int32(value), err
}

func (v *Value) Int64() (int64, error) {
	if ret, ok := v.value.(int64); ok {
		return ret, nil
	}
	value, err := strconv.ParseInt(v.String(), 10, 64)
	return int64(value), err
}

func (v *Value) Uint() (uint, error) {
	if ret, ok := v.value.(uint); ok {
		return ret, nil
	}
	value, err := strconv.ParseUint(v.String(), 10, 32)
	return uint(value), err
}

func (v *Value) Uint8() (uint8, error) {
	if ret, ok := v.value.(uint8); ok {
		return ret, nil
	}
	value, err := strconv.ParseUint(v.String(), 10, 8)
	return uint8(value), err
}

func (v *Value) Uint16() (uint16, error) {
	if ret, ok := v.value.(uint16); ok {
		return ret, nil
	}
	value, err := strconv.ParseUint(v.String(), 10, 16)
	return uint16(value), err
}

func (v *Value) Uint32() (uint32, error) {
	if ret, ok := v.value.(uint32); ok {
		return ret, nil
	}
	value, err := strconv.ParseUint(v.String(), 10, 32)
	return uint32(value), err
}

func (v *Value) Uint64() (uint64, error) {
	if ret, ok := v.value.(uint64); ok {
		return ret, nil
	}
	value, err := strconv.ParseUint(v.String(), 10, 64)
	return uint64(value), err
}

func (v *Value) String() string {
	if ret, ok := v.value.(string); ok {
		return ret
	}
	if v.IsNil() {
		return ""
	}
	return ToStr(v.value)
}

func (v *Value) MustBool() (ret bool) {
	ret, _ = v.Bool()
	return
}

func (v *Value) MustFloat32() (ret float32) {
	ret, _ = v.Float32()
	return
}

func (v *Value) MustFloat64() (ret float64) {
	ret, _ = v.Float64()
	return
}

func (v *Value) MustInt() (ret int) {
	ret, _ = v.Int()
	return
}

func (v *Value) MustInt8() (ret int8) {
	ret, _ = v.Int8()
	return
}

func (v *Value) MustInt16() (ret int16) {
	ret, _ = v.Int16()
	return
}

func (v *Value) MustInt32() (ret int32) {
	ret, _ = v.Int32()
	return
}

func (v *Value) MustInt64() (ret int64) {
	ret, _ = v.Int64()
	return
}

func (v *Value) MustUint() (ret uint) {
	ret, _ = v.Uint()
	return
}

func (v *Value) MustUint8() (ret uint8) {
	ret, _ = v.Uint8()
	return
}

func (v *Value) MustUint16() (ret uint16) {
	ret, _ = v.Uint16()
	return
}

func (v *Value) MustUint32() (ret uint32) {
	ret, _ = v.Uint32()
	return
}

func (v *Value) MustUint64() (ret uint64) {
	ret, _ = v.Uint64()
	return
}

func (v *Value) MustString() (ret string) {
	ret = v.String()
	return
}

func (v *Value) MustSliceString() (ret []string) {
	ret, _ = v.value.([]string)
	return
}

func (v *Value) MustBytes() (ret []byte) {
	ret, _ = v.value.([]byte)
	return
}

func (v *Value) Value() interface{} {
	return v.value
}

func (v *Value) IsNil() bool {
	return v.value == nil
}

func ValueTo(value interface{}) *Value {
	return &Value{value: value}
}
