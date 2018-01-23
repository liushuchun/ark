package urlencode

import (
	"testing"
)

func TestEscape(t *testing.T) {

	escapes := map[string]string{
		"/":        "@/",
		"@@":       "@@@",
		"//":       "@/@/",
		"a//":      "a/@/",
		"///":      "@/@/@/",
		"a/@b":     "a/@@b",
		"a/////":   "a/@/@/@/@/",
		"/a//b//c": "@/a/@/b/@/c",
		"@/":       "@@/",
		"@":        "@@",
		"//@":      "@/@/@@",
		//以下是不需要转义的测试案例
		"1234/56":   "1234/56",
		"123@/56":   "123@/56",
		"123456":    "123456",
		"4/5/6/7/7": "4/5/6/7/7",
		"4@/5@/":    "4@/5@/",
		"a@@@":      "a@@@",
	}

	for k, v := range escapes {
		if Escape(k) != v {
			t.Fatalf("escape failed: %s  expect:%s but:%s \n", k, v, Escape(k))
		}
		if UnEscape(v) != k {
			t.Fatalf("unescape failed: %s  expect:%s but:%s \n", v, k, UnEscape(v))
		}
	}
}

func TestUrlEncoding(t *testing.T) {

	escapes := map[string]string{
		"/":        "%40%2F",
		"@@":       "%40%40%40",
		"//":       "%40%2F%40%2F",
		"a//":      "a%2F%40%2F",
		"///":      "%40%2F%40%2F%40%2F",
		"a/@b":     "a%2F%40%40b",
		"a/////":   "a%2F%40%2F%40%2F%40%2F%40%2F",
		"/a//b//c": "%40%2Fa%2F%40%2Fb%2F%40%2Fc",
		"@/":       "%40%40%2F",
		"@":        "%40%40",
		"//@":      "%40%2F%40%2F%40%40",
	}
	for k, v := range escapes {
		if UrlEncoding(k) != v {
			t.Fatalf("UrlEncoding failed: %s  expect:%s but:%s \n", k, v, UrlEncoding(k))
		}
		if key, _ := UrlUnEncoding(v); key != k {
			t.Fatalf("UrlUnEncoding failed: %s  expect:%s but:%s \n", v, k, key)
		}
	}
}
