package tab

import (
	. "github.com/qiniu/ctype"
	"syscall"
)

const (
	TABLECH   = TSPACE | BLANK | RDIV
	UNTABLECH = LCAP_W | LCAP_R | LCAP_N | LCAP_T | RDIV
)

// -----------------------------------------------------------
// func Escape

var g_escape = []byte{
	'w',  //   [0]
	't',  //   [1]
	'n',  //   [2]
	0,    //   [3]
	'\\', //   [4]
	'r',  //   [5]
}

//
// 转义符：9(\t), 10(\n), 13(\r), 32(\w), 92(\\)
// %8: 1, 2, 5, 0, 4
//
func Escape(text string) string {

	nesc := needEscape(text)

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

func needEscape(text string) int {

	nesc := 0
	for i := 0; i < len(text); i++ {
		if Is(TABLECH, rune(text[i])) {
			nesc++
		}
	}
	return nesc
}

// -----------------------------------------------------------
// func Unescape

var g_unescape = []byte{
	' ',  //   [0]
	'\\', //   [1]
	'\r', //   [2]
	0,    //   [3]
	'\t', //   [4]
	'\n', //   [5]
}

//
// 转义符：116(t), 110(n), 114(r), 92(\), 119(w)
// %7: 4, 5, 2, 1, 0
//
func Unescape(text2 string) (text string, err error) {

	nesc := 0
	n := len(text2) - 1
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
	if nesc == 0 {
		return text2, nil
	}

	text1 := make([]byte, len(text2)-nesc)
	i1 := 0
	i = 0
	for i < n {
		c := text2[i]
		if c == '\\' {
			text1[i1] = g_unescape[text2[i+1]%7]
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
