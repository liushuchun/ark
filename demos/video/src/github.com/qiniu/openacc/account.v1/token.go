package account

import (
	"crypto/sha1"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"strings"

	. "github.com/qiniu/openacc/account.api.v1"
	"qiniupkg.com/http/httputil.v2"
)

var (
	ErrBadToken        = httputil.NewError(401, "bad token")
	ErrInvalidKeyPairs = httputil.NewError(400, "invalid keypairs")
)

// ---------------------------------------------------------------------------

type KeyPair struct {
	AccessKey string `json:"ak"`
	SecretKey string `json:"sk"`
}

type Config struct {
	KeyPairs []KeyPair `json:"keypairs"`
}

// ---------------------------------------------------------------------------

type keyPairImpl struct{
	AccessKey string
	SecretKey []byte
}

type Manager struct {
	keyPairs []keyPairImpl
}

func New(cfg *Config) (p *Manager, err error) {

	n := len(cfg.KeyPairs)
	if n < 1 || n > 9 {
		err = ErrInvalidKeyPairs
		return
	}

	keyPairs := make([]keyPairImpl, n)
	for i, pair := range cfg.KeyPairs {
		keyPairs[i] = keyPairImpl{pair.AccessKey, []byte(pair.SecretKey)}
	}

	p = &Manager{
		keyPairs: keyPairs,
	}
	return
}

// token = <AccessKey>:base64(<Sign><Data>)
//
func (p *Manager) MakeToken(user *UserInfo) string {

	data, _ := json.Marshal(user)
	pair := p.keyPairs[0]
	h := hmac.New(sha1.New, pair.SecretKey)
	h.Write(data)
	sign := h.Sum(nil)
	token := pair.AccessKey + ":" + base64.URLEncoding.EncodeToString(append(sign, data...))
	return token
}

func (p *Manager) ParseToken(token string) (user *UserInfo, err error) {

	pos := strings.Index(token, ":")
	if pos <= 0 {
		err = ErrBadToken
		return
	}

	b, err := base64.URLEncoding.DecodeString(token[pos+1:])
	if err != nil {
		return
	}

	ak := token[:pos]
	for _, pair := range p.keyPairs {
		if pair.AccessKey == ak {
			realMAC := b[:20]
			data := b[20:]
			h := hmac.New(sha1.New, pair.SecretKey)
			h.Write(data)
			expectedMAC := h.Sum(nil)
			if hmac.Equal(realMAC, expectedMAC) {
				user = new(UserInfo)
				err = json.Unmarshal(data, user)
				if err == nil {
					return
				}
			}
			break
		}
	}
	err = ErrBadToken
	return
}

// ---------------------------------------------------------------------------

