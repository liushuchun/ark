package seshcookie

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"github.com/qiniu/log.v1"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------

func encodeGob(obj interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func decodeGob(encoded []byte) (map[string]interface{}, error) {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	var out map[string]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func encodeCookie(content interface{}, key, iv []byte) (string, error) {
	sessionGob, err := encodeGob(content)
	if err != nil {
		return "", err
	}
	padLen := aes.BlockSize - (len(sessionGob)+4)%aes.BlockSize
	buf := bytes.NewBuffer(nil)
	var sessionLen int32 = (int32)(len(sessionGob))
	binary.Write(buf, binary.BigEndian, sessionLen)
	buf.WriteString(sessionGob)
	buf.WriteString(strings.Repeat("\000", padLen))
	sessionBytes := buf.Bytes()
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	encrypter := cipher.NewCBCEncrypter(aesCipher, iv)
	encrypter.CryptBlocks(sessionBytes, sessionBytes)
	b64 := base64.URLEncoding.EncodeToString(sessionBytes)
	return b64, nil
}

func decodeCookie(encodedCookie string, key, iv []byte) (map[string]interface{}, error) {
	sessionBytes, err := base64.URLEncoding.DecodeString(encodedCookie)
	if err != nil {
		log.Printf("base64.Decodestring: %s\n", err)
		return nil, err
	}
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("aes.NewCipher: %s\n", err)
		return nil, err
	}
	// decrypt in-place
	decrypter := cipher.NewCBCDecrypter(aesCipher, iv)
	decrypter.CryptBlocks(sessionBytes, sessionBytes)

	buf := bytes.NewBuffer(sessionBytes)
	var gobLen int32
	binary.Read(buf, binary.BigEndian, &gobLen)
	gobBytes := sessionBytes[4 : 4+gobLen]
	session, err := decodeGob(gobBytes)
	if err != nil {
		log.Printf("decodeGob: %s\n", err)
		return nil, err
	}
	return session, nil
}

// -----------------------------------------------------------

type sessionWriter struct {
	http.ResponseWriter
	session     map[string]interface{}
	h           *SessionManager
	req         *http.Request
	wroteHeader int32
}

func (s *sessionWriter) WriteHeader(code int) {

	if atomic.AddInt32(&s.wroteHeader, 1) == 1 {
		origCookie, err := s.req.Cookie(s.h.CookieName)
		var origCookieVal string
		if err != nil {
			origCookieVal = ""
		} else {
			origCookieVal = origCookie.Value
		}

		session := s.session
		if len(session) == 0 {
			// if we have an empty session, but the
			// request didn't start out that way, we
			// assume the user wants us to clear the
			// session
			if origCookieVal != "" {
				log.Println("clearing cookie")
				var cookie http.Cookie
				cookie.Name = s.h.CookieName
				if s.h.CookieDomain != "" {
					cookie.Domain = s.h.CookieDomain
				}
				cookie.Value = ""
				cookie.Path = "/"
				// a cookie is expired by setting it
				// with an expiration time in the past
				cookie.Expires = time.Now().Add(-3600)
				http.SetCookie(s, &cookie)
			} else {
				log.Println("not setting empty cookie")
			}
			goto write
		}
		encoded, err := encodeCookie(session, s.h.key, s.h.iv)
		if err != nil {
			log.Printf("createCookie: %s\n", err)
			goto write
		}

		if encoded == origCookieVal {
			//log.Println("not re-setting identical cookie")
			goto write
		}

		var cookie http.Cookie
		cookie.Name = s.h.CookieName
		if s.h.CookieDomain != "" {
			cookie.Domain = s.h.CookieDomain
		}
		cookie.Value = encoded
		cookie.Path = "/"
		http.SetCookie(s, &cookie)
	}
write:
	s.ResponseWriter.WriteHeader(code)
}

func (s *sessionWriter) Write(body []byte) (int, error) {
	if s.wroteHeader == 0 {
		s.WriteHeader(http.StatusOK)
	}
	return s.ResponseWriter.Write(body)
}

// -----------------------------------------------------------

type SessionManager struct {
	// The name of the cookie our encoded session will be stored
	// in.
	CookieName   string
	CookieDomain string
	key          []byte
	iv           []byte
}

func NewSessionManager(cookieName, cookieDomain, key string) *SessionManager {
	// sha1 sums are 20 bytes long.  we use the first 16 bytes as
	// the aes key, and the last 16 bytes as the initialization
	// vector (understanding that they overlap, of course).
	keySha1 := sha1.New()
	keySha1.Write([]byte(key))
	sum := keySha1.Sum(nil)
	return &SessionManager{
		CookieName:   cookieName,
		CookieDomain: cookieDomain,
		key:          sum[:16],
		iv:           sum[4:],
	}
}

func (h *SessionManager) getCookieSession(req *http.Request) map[string]interface{} {
	cookie, err := req.Cookie(h.CookieName)
	if err != nil {
		return map[string]interface{}{}
	}
	session, err := decodeCookie(cookie.Value, h.key, h.iv)
	if err != nil {
		log.Warn("decodeCookie: %s\n", err)
		return map[string]interface{}{}
	}

	return session
}

func (h *SessionManager) Get(rw http.ResponseWriter, req *http.Request) (http.ResponseWriter, map[string]interface{}) {
	// get our session a little early, so that we can add our
	// authentication information to it if we get some
	session := h.getCookieSession(req)
	sessionWriter := &sessionWriter{rw, session, h, req, 0}
	return sessionWriter, session
}

// -----------------------------------------------------------
