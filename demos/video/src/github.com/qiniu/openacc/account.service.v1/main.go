package account

import (
	"net/http"
	"strings"

	"github.com/qiniu/openacc/account.v1"
	"qiniupkg.com/http/httputil.v2"

	api "github.com/qiniu/openacc/account.api.v1"
	. "github.com/qiniu/openacc/storage/proto.v1"
)

var (
	ErrPostMethodOnly     = httputil.NewError(400, "only allow POST method")
	ErrBadUsernameOrPwd   = httputil.NewError(401, "bad username or password")
	ErrBadToken           = httputil.NewError(401, "bad token")
	ErrExpiredToken       = httputil.NewError(401, "expired token")
	ErrAccessDenied       = httputil.NewError(403, "access denied")
	ErrInvalidEmailStatus = httputil.NewError(400, "email_status is invalid: must be an integer value")
	ErrInvalidUtype       = httputil.NewError(400, "utype is invalid: must be an integer value")
	ErrNotImpl            = httputil.NewError(599, "not impl")
)

const (
	authMethod = "Bearer "
)

// ---------------------------------------------------------------------------

type Config struct {
	Storage      Storage
	KeyPairs     []account.KeyPair
	SuperAdminId string
}

type Service struct {
	Config
	tokens *account.Manager
}

func New(cfg *Config) (p *Service, err error) {

	p = &Service{
		Config: *cfg,
	}
	p.tokens, err = account.New(&account.Config{
		KeyPairs: cfg.KeyPairs,
	})
	return
}

func (p *Service) canSetUtype(admin *api.UserInfo, utype int) (err error) {

	if admin.Id == p.SuperAdminId {
		return
	}

	if (utype & api.UtypeAdmin) != 0 {
		return ErrAccessDenied
	}
	return
}

func (p *Service) validateAdminToken(req *http.Request) (user *api.UserInfo, err error) {

	user, err = p.validateToken(req)
	if err != nil {
		return
	}
	if user.IsAdmin() {
		return
	}
	return user, ErrBadToken
}

func (p *Service) validateToken(req *http.Request) (user *api.UserInfo, err error) {

	auth := req.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, authMethod) {
		return nil, ErrBadToken
	}

	token := auth[len(authMethod):]
	user, err = p.tokens.ParseToken(token)
	if err != nil {
		return
	}
	if user.IsValid() {
		return
	}
	return user, ErrExpiredToken
}

func (p *Service) makeToken(id string, entry *UserEntry, req *http.Request, expiry int64) string {

	return p.tokens.MakeToken(&api.UserInfo{
		Id:     id,
		Name:   entry.Name,
		Appid:  req.Form.Get("client_id"),
		Devid:  req.Form.Get("device_id"),
		Expiry: expiry,
		Utype:  int(entry.Utype),
	})
}

// ---------------------------------------------------------------------------

