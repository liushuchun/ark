package strconv

import (
	"strconv"
)

func Atoui(s string) (i uint, err error) {
	i64, err := strconv.ParseUint(s, 10, 0)
	return uint(i64), err
}

func Uitoa(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}
