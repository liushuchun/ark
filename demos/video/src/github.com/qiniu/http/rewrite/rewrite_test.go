package rewrite

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// ------------------------------------------------------------------------

func TestBasic(t *testing.T) {

	re := regexp.MustCompile("(^a(x*)b$)|(^a(.*)b$)")
	fmt.Printf("%q\n", re.FindStringSubmatch("ab"))
	fmt.Printf("%q\n", re.FindStringSubmatch("axxb"))
	fmt.Printf("%q\n", re.FindStringSubmatch("ab-axb"))
	fmt.Printf("%q\n", re.FindStringSubmatch("axxb-ab"))
}

func TestBasic2(t *testing.T) {

	re := regexp.MustCompile("(^a(x*)b$)|(^a(.*)b$)")
	fmt.Printf("%v\n", re.FindStringSubmatchIndex("ab"))
	fmt.Printf("%v\n", re.FindStringSubmatchIndex("axxb"))
	fmt.Printf("%v\n", re.FindStringSubmatchIndex("ab-axb"))
	fmt.Printf("%v\n", re.FindStringSubmatchIndex("axxb-ab"))
}

func TestSubexpNames(t *testing.T) {

	text := "(?P<named>^a(x*)b$)|(?:^a(.*)b$)"
	re := regexp.MustCompile(text)
	subs := re.SubexpNames()
	fmt.Printf("SubexpNames: %q - %d, %d\n", subs, re.NumSubexp(), strings.Count(text, "("))
}

// ------------------------------------------------------------------------

func TestRewrite(t *testing.T) {

	items := []RouteItem{
		{"a(x*)b", "b${1}a"},
		{"a(.*)b", "B$1/A"},
	}
	p, err := Compile(items)
	if err != nil {
		t.Fatal("Compile failed:", err)
	}

	dest, err := p.Rewrite("axxb")
	fmt.Println(dest, err)
	if err != nil || dest != "bxxa" {
		t.Fatal("Rewrite failed:", err)
	}

	dest, err = p.Rewrite("ab-axb")
	fmt.Println(dest, err)
	if err != nil || dest != "Bb-ax/A" {
		t.Fatal("Rewrite failed:", err)
	}
}

func TestRewrite2(t *testing.T) {

	items := []RouteItem{
		{"a(.*)b", "B$1/A"},
		{"a(x*)b(y*)", "${1}${2}"},
	}

	p, err := Compile(items)
	if err != nil {
		t.Fatal("Compile failed:", err)
	}

	dest, err := p.Rewrite("axxbyy")
	fmt.Println(dest, err)
	if err != nil || dest != "xxyy" {
		t.Fatal("Rewrite failed:", err)
	}
}

// ------------------------------------------------------------------------
