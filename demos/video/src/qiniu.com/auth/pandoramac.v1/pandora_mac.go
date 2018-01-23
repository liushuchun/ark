package pandoramac

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/qiniu/http/httputil.v1"
)

var (
	ErrInvalidDateFormat    = httputil.NewError(400, "Format of date in request must be GMT")
	ErrRequestTimeTooSkewed = httputil.NewError(403, "Request time differ with server time must be less than 15 minutes")
	ErrMissingDateHeader    = httputil.NewError(400, "The request must include the Date in the header")
)

const qiniuHeaderPrefix = "X-Qiniu-"
const RequestTimeSkewedLimit = 900 // 900s

var qiniuSubResource = []string{}

func SignQiniuHeaderValues(header http.Header) (out string) {
	var keys []string
	for key, _ := range header {
		if len(key) > len(qiniuHeaderPrefix) && key[:len(qiniuHeaderPrefix)] == qiniuHeaderPrefix {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return
	}

	if len(keys) > 1 {
		sort.Sort(headerKey(keys))
	}
	for _, key := range keys {
		out += fmt.Sprintf("\n%s:%s", strings.ToLower(key), header.Get(key))
	}
	return
}

func SignQiniuResourceValues(u *url.URL) (out string) {
	out += u.Path

	var keys []string
	query := u.Query()
	for _, v := range qiniuSubResource {
		if query.Get(v) != "" {
			keys = append(keys, query.Get(v))
		}
	}
	if len(keys) == 0 {
		return
	}

	for i, k := range keys {
		if i == 0 {
			out += "?"
		}
		out += fmt.Sprintf("%s=%s", k, query.Get(k))
		if i != len(keys)-1 {
			out += "&"
		}
	}
	return
}

func SignRequest(sk []byte, req *http.Request) ([]byte, error) {
	if err := checkRequest(req); err != nil {
		return nil, err
	}

	h := hmac.New(sha1.New, sk)

	io.WriteString(h,
		fmt.Sprintf("%s\n%s\n%s\n%s\n",
			req.Method,
			req.Header.Get("Content-MD5"),
			req.Header.Get("Content-Type"),
			req.Header.Get("Date")))

	io.WriteString(h, SignQiniuHeaderValues(req.Header))
	io.WriteString(h, SignQiniuResourceValues(req.URL))

	return h.Sum(nil), nil
}

func SignAdminRequest(sk []byte, req *http.Request, su string) ([]byte, error) {
	if err := checkRequest(req); err != nil {
		return nil, err
	}

	h := hmac.New(sha1.New, sk)

	io.WriteString(h,
		fmt.Sprintf("%s\n%s\n%s\n%s\nAuthorization: PandoraAdmin %s",
			req.Method,
			req.Header.Get("Content-MD5"),
			req.Header.Get("Content-Type"),
			req.Header.Get("Date"),
			su))

	io.WriteString(h, SignQiniuHeaderValues(req.Header))
	io.WriteString(h, SignQiniuResourceValues(req.URL))

	return h.Sum(nil), nil
}

func checkRequest(req *http.Request) error {
	// check Date
	date := req.Header.Get("Date")
	if date == "" {
		return ErrMissingDateHeader
	}
	t, err := time.Parse(http.TimeFormat, date)
	if err != nil {
		return ErrInvalidDateFormat
	}
	curr := time.Now()
	if curr.Unix()-t.Unix() > RequestTimeSkewedLimit || t.Unix()-curr.Unix() > RequestTimeSkewedLimit {
		return ErrRequestTimeTooSkewed
	}

	return nil
}

type RequestSigner struct {
}

var (
	DefaultRequestSigner RequestSigner
)

func (p RequestSigner) Sign(sk []byte, req *http.Request) ([]byte, error) {
	return SignRequest(sk, req)
}

func (p RequestSigner) SignAdmin(sk []byte, req *http.Request, su string) ([]byte, error) {
	return SignAdminRequest(sk, req, su)
}

type headerKey []string

func (p headerKey) Len() int           { return len(p) }
func (p headerKey) Less(i, j int) bool { return p[i] < p[j] }
func (p headerKey) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
