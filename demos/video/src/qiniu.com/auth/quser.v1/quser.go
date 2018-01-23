package quser

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"

	"github.com/qiniu/bytes/seekable"
)

// ---------------------------------------------------------------------------------------

func incBody(req *http.Request, ctType string) bool {

	return req.ContentLength != 0 && req.Body != nil && ctType != "" && ctType != "application/octet-stream"
}

func SignRequest(sk []byte, req *http.Request, quserAk string) ([]byte, error) {

	h := hmac.New(sha1.New, sk)

	u := req.URL
	data := req.Method + " " + u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\nHost: "+req.Host)

	ctType := req.Header.Get("Content-Type")
	if ctType != "" {
		io.WriteString(h, "\nContent-Type: "+ctType)
	}
	io.WriteString(h, "\nAuthorization: QUser "+quserAk+"\n\n")

	if incBody(req, ctType) {
		s2, err2 := seekable.New(req)
		if err2 != nil {
			return nil, err2
		}
		h.Write(s2.Bytes())
	}

	return h.Sum(nil), nil
}

func SignAdminRequest(sk []byte, req *http.Request, su, quserAk string) ([]byte, error) {

	h := hmac.New(sha1.New, sk)

	u := req.URL
	data := req.Method + " " + u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\nHost: "+req.Host)

	ctType := req.Header.Get("Content-Type")
	if ctType != "" {
		io.WriteString(h, "\nContent-Type: "+ctType)
	}
	fmt.Fprintf(h, "\nAuthorization: QAdmin %s:%s\n\n", su, quserAk)

	if incBody(req, ctType) {
		s2, err2 := seekable.New(req)
		if err2 != nil {
			return nil, err2
		}
		h.Write(s2.Bytes())
	}

	return h.Sum(nil), nil
}

// ---------------------------------------------------------------------------------------
