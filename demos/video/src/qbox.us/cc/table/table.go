package table

import (
	. "github.com/qiniu/ctype"
	"syscall"
)

const (
	TABLECH   = TSPACE | RDIV
	UNTABLECH = LCAP_R | LCAP_N | LCAP_T | RDIV
)

// -----------------------------------------------------------
// func Escape

var g_escape = []byte{
	0,    //   [0]
	't',  //   [1]
	'n',  //   [2]
	0,    //   [3]
	'\\', //   [4]
	'r',  //   [5]
}

//
// 转义符：9(\t), 10(\n), 13(\r), 92(\\)
// %8: 1, 2, 5, 4
//
// 本 package 已经过时，建议用 github.com/qiniu/encoding/tab 包
//
func Escape(text string) string {

	nesc := 0
	for i := 0; i < len(text); i++ {
		if Is(TABLECH, rune(text[i])) {
			nesc++
		}
	}

	if nesc == 0 {
		return text
	}
	text2 := make([]byte, len(text)+nesc)
	i2 := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		if Is(TABLECH, rune(c)) {
			text2[i2] = '\\'
			i2++
			text2[i2] = g_escape[c&7]
		} else {
			text2[i2] = c
		}
		i2++
	}
	return string(text2)
}

// -----------------------------------------------------------
// func Unescape

var g_unescape = []byte{
	'\n', //   [0]
	'\t', //   [1]
	'\\', //   [2]
	0,    //   [3]
	'\r', //   [4]
	0,    //   [5]
}

//
// 转义符：116(t), 110(n), 114(r), 92(\)
// %5: 1, 0, 4, 2
//
func Unescape(text2 string) (text string, err error) {

	n := len(text2) - 1
	nesc := -1
	i := 0
	for i < n {
		if text2[i] == '\\' {
			if Is(UNTABLECH, rune(text2[i+1])) {
				nesc++
				i++
			} else {
				err = syscall.EINVAL
				return
			}
		}
		i++
	}

	if nesc < 0 {
		return text2, nil
	}

	text1 := make([]byte, n-nesc)
	i1 := 0
	i = 0
	for i < n {
		c := text2[i]
		if c == '\\' {
			text1[i1] = g_unescape[text2[i+1]%5]
			i++
		} else {
			text1[i1] = c
		}
		i++
		i1++
	}
	if i == n {
		text1[i1] = text2[i]
	}
	return string(text1), nil
}

// -----------------------------------------------------------
