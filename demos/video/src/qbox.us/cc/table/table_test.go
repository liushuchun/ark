package table

import (
	"github.com/qiniu/ts"
	"testing"
)

type escapeCase struct {
	text   string
	result string
}

var g_escapeCases = []escapeCase{
	{"a bc", "a bc"},
	{"a b\\c", "a b\\\\c"},
	{"a bc\\", "a bc\\\\"},
	{"\ta bc", "\\ta bc"},
	{"\ta bc\n", "\\ta bc\\n"},
}

func TestEscape(t *testing.T) {

	for _, c := range g_escapeCases {
		text2 := Escape(c.text)
		if text2 != c.result {
			ts.Fatalf(t, "Escape failed: '%s' '%s' '%s'\n", c.text, c.result, text2)
		}
		text1, err := Unescape(c.result)
		if err != nil || c.text != text1 {
			ts.Fatalf(t, "Unescape failed: '%s' '%s' '%s' %v %v %v\n", c.result, c.text, text1, len(c.text), len(text1), err)
		}
	}
}

type unescapeCase struct {
	text   string
	result string
	ok     bool
}

var g_unescapeCases = []unescapeCase{
	{"a \\bc", "a bc", false},
	{"a bc\\", "a bc\\", true},
}

func TestUnescape(t *testing.T) {

	for _, c := range g_unescapeCases {
		text1, err := Unescape(c.text)
		if err != nil {
			if c.ok {
				ts.Fatal(t, "Unescape failed:", err)
			}
		} else {
			if text1 != c.result {
				ts.Fatalf(t, "Unescape failed: '%s' '%s' '%s'\n", c.text, c.result, text1)
			}
		}
	}
}
