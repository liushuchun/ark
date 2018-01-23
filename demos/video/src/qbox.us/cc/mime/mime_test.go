package mime

import (
	"testing"

	"github.com/stretchr/testify.v1/require"
	"github.com/stretchr/testify/assert"

	"github.com/qiniu/ts"
)

type T struct {
	FileName string
	Expect   string
}

func TestMime(t *testing.T) {

	cases := []T{
		{".txt", "text/plain"},
		{".m4a", "audio/mpeg"},
		{".aac", "audio/x-aac"},
		{".webm", "video/webm"},
		{".IPA", "application/octet-stream"},
		{".webp", "image/webp"},
		{".mp3", "audio/mpeg"},
	}

	for _, c := range cases {
		if c.Expect != ContentType(c.FileName) {
			t.Fatalf("%v expect %v but %v \n",
				c.FileName,
				c.Expect,
				ContentType(c.FileName))
		}
	}
}

func TestRMime(t *testing.T) {

	cases := []T{
		{"", "application/octet-stream"},
		{".m3u8", "application/vnd.apple.mpegurl"},
	}

	for _, c := range cases {
		assert.Equal(t, c.FileName, Suffix(c.Expect))
	}

	c := T{".jpg", "image/jpeg"}
	for i := 0; i < 10; i++ {
		generateRmimes()
		assert.Equal(t, c.FileName, Suffix(c.Expect))
	}
}

func TestIsValidMimeType(t *testing.T) {
	cases := []struct {
		mime   string
		result bool
	}{
		{"image/jpeg", true},
		{"video/vnd.fvt", true},
		{"video/x-flv", true},
		{"application/atom+xml", true},
		{"text/plain; charset=iso-8859-1", true},
		{"", true},
		{"\n", false},
		{"\r", false},
		{"\t", false},
		{"%0A", false},
		{"%0D", false},
		{"%09", false},
	}

	for _, v := range cases {
		if IsValidMimeType(v.mime) != v.result {
			ts.Fatal(t, "Check Avaliable Mime failed:", v.mime, v.result)
		}
	}
	s := make([]byte, 201)
	for i := 0; i < 201; i++ {
		s[i] = 'A'
	}
	if IsValidMimeType(string(s)) != false {
		ts.Fatal(t, "Check Avaliable Mime failed:", string(s))
	}
}

func TestMetaKey(t *testing.T) {
	cases := map[string]bool{
		"":   true,
		"a":  true,
		"-":  true,
		"_":  true,
		"1":  true,
		"'":  false,
		".":  false,
		"/":  false,
		"+":  false,
		"=":  false,
		"\n": false,
		"\t": false,
		"123456789012345678901234567890123456789012345678901": false,
	}
	for k, v := range cases {
		require.Equal(t, v, IsValidMetaKey("x-qn-meta-"+k), k)
	}
}
