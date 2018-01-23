package pandoramac

import (
	"crypto/hmac"
	"crypto/sha1"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify.v1/assert"
)

var (
	sk = []byte("secret_key")
	su = "su_info"
)

func Test_Sign(t *testing.T) {
	date := time.Now().UTC().Format(http.TimeFormat)
	req, err := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-MD5", "xxx")
	req.Header.Set("Date", date)

	act, err := SignRequest(sk, req)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET\nxxx\napplication/json\n" + date +
		"\n/path/to/api"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_SignWithXQiniu(t *testing.T) {
	date := time.Now().UTC().Format(http.TimeFormat)
	req, err := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Qiniu-Meta-App", "value")
	req.Header.Set("Date", date)

	act, err := SignRequest(sk, req)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET\n\napplication/json\n" + date +
		"\n\nx-qiniu-meta-app:value/path/to/api"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_SignAdmin(t *testing.T) {
	date := time.Now().UTC().Format(http.TimeFormat)
	req, err := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Date", date)

	act, err := SignAdminRequest(sk, req, su)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET\n\napplication/json\n" + date +
		"\nAuthorization: PandoraAdmin " + su +
		"/path/to/api"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_SignAdminWithXQiniu(t *testing.T) {
	date := time.Now().UTC().Format(http.TimeFormat)
	req, err := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Qiniu-Meta-App", "value")
	req.Header.Set("Date", date)

	act, err := SignAdminRequest(sk, req, su)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET\n\napplication/json\n" + date +
		"\nAuthorization: PandoraAdmin " + su +
		"\nx-qiniu-meta-app:value" +
		"/path/to/api"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_signQiniuHeaderValues(t *testing.T) {
	header := make(http.Header)
	header.Set("X-Qbox-Meta", "value")

	assert.Empty(t, SignQiniuHeaderValues(header))

	header.Set("X-Qiniu-Cxxxx", "valuec")
	header.Set("X-Qiniu-Bxxxx", "valueb")
	header.Set("X-Qiniu-axxxx", "valuea")
	header.Set("X-Qiniu-e", "value")
	header.Set("X-Qiniu-", "value")
	header.Set("X-Qiniu", "value")
	header.Set("", "value")

	assert.Equal(t, "\nx-qiniu-axxxx:valuea\nx-qiniu-bxxxx:valueb\nx-qiniu-cxxxx:valuec\nx-qiniu-e:value", SignQiniuHeaderValues(header))
}
