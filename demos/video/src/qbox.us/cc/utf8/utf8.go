package utf8

import (
	"unicode/utf8"
)

func ValidUtf8(intf interface{}) bool {

	switch val := intf.(type) {
	case string:
		return ValidString(val)
	case map[string]string:
		for k, v := range val {
			if !ValidString(k) || !ValidString(v) {
				return false
			}
		}
		return true
	}
	return true
}

// 只要字符串中出现 utf8.RuneError 就认为是非utf8字符串
func ValidString(s string) bool {

	for _, r := range s {
		if r == utf8.RuneError {
			return false
		}
	}
	return true
}
