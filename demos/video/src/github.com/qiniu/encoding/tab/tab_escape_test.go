package tab

import (
	"github.com/qiniu/ts"
	"testing"
)

// -----------------------------------------------------------

var cases = [][2]string{
	{"a \tb", "a\\w\\tb"},
	{"a\n \t\rb", "a\\n\\w\\t\\rb"},
}

func Test(t *testing.T) {

	for _, c := range cases {
		esc := Escape(c[0])
		if esc != c[1] {
			ts.Fatal(t, "Escape failed:", esc, c[1])
		}
		unesc, err := Unescape(esc)
		if err != nil || unesc != c[0] {
			ts.Fatal(t, "Unescape failed:", unesc, c[0], err)
		}
	}
}

func Tests(t *testing.T) {

	a := []interface{}{"hello world!\r\n", 124}
	esc := Escapes(a)
	if len(esc) != 2 {
		ts.Fatal(t, "Escapes failed:", len(esc))
	}
	if esc[0].(string) != "hello\\wworld!\\r\\n" {
		ts.Fatal(t, "Escapes esc[0]:", esc[0])
	}
	if esc[1].(int) != 124 {
		ts.Fatal(t, "Escapes esc[1]:", esc[1])
	}

	var v1 string
	var v2 int
	var unesc = []interface{}{&v1, &v2}
	v1 = esc[0].(string)
	v2 = esc[1].(int)
	err := Unescapes(unesc)
	if err != nil || v1 != a[0].(string) || v2 != a[0].(int) {
		ts.Fatal(t, "Unescapes failed:", unesc, err)
	}
}

// -----------------------------------------------------------
