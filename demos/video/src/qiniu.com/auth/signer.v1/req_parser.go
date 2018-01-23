package signer

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"

	. "code.google.com/p/go.net/context"
	"qiniu.com/auth/authutil.v1"
	. "qiniu.com/auth/proto.v1"
)

var (
	ErrBadToken = httputil.NewError(401, "bad token")
)

// --------------------------------------------------------------------

type requestSigner interface {
	Sign(sk []byte, req *http.Request) ([]byte, error)
	SignAdmin(sk []byte, req *http.Request, su string) ([]byte, error)
}

func ParseNormalAuth(
	rs requestSigner, acc Interface, ctx Context,
	token string, req *http.Request) (user SudoerInfo, err error) {

	pos := strings.Index(token, ":")
	if pos == -1 {
		err = ErrBadToken
		return
	}

	key := token[:pos]

	info, err := acc.GetAccessInfo(ctx, key)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseAuth: GetAccessInfo").Detail(err)
		return
	}

	sign, err := rs.Sign(info.Secret, req)
	if err != nil {
		err = errors.Info(err, "parseAuth: SignRequest").Detail(err)
		return
	}

	signExp := token[pos+1:]
	if base64.URLEncoding.EncodeToString(sign) != signExp {
		err = errors.Info(ErrBadToken, "parseAuth: checksum error")
		return
	}

	user.Access = key
	user.Appid = info.Appid
	user.Uid = info.Uid
	user.Utype, err = acc.GetUtype(ctx, user.Uid)
	return
}

func ParseAdminAuth(
	rs requestSigner, acc Interface, ctx Context,
	token string, req *http.Request) (user SudoerInfo, err error) {

	pos := strings.Index(token, ":")
	if pos == -1 {
		err = ErrBadToken
		return
	}

	suInfo := token[:pos]
	token = token[pos+1:]

	uid, appid, err := authutil.ParseSuInfo(suInfo)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseSuInfo: ", suInfo).Detail(err)
		return
	}

	pos = strings.Index(token, ":")
	if pos == -1 {
		err = ErrBadToken
		return
	}

	key := token[:pos]
	signExp := token[pos+1:]

	info, err := acc.GetAccessInfo(ctx, key)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseAdminAuth: GetAccessInfo").Detail(err)
		return
	}
	utypeSu, err := acc.GetUtype(ctx, info.Uid)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseAdminAuth: GetUtypeSu").Detail(err)
		return
	}
	if (utypeSu & USER_TYPE_SUDOERS) == 0 {
		err = errors.Info(ErrBadToken, "parseAdminAuth: not sudoer").Detail(err)
		return
	}

	sign, err := rs.SignAdmin(info.Secret, req, suInfo)
	if err != nil {
		err = errors.Info(err, "parseAdminAuth: SignAdminRequest").Detail(err)
		return
	}
	if base64.URLEncoding.EncodeToString(sign) != signExp {
		err = errors.Info(ErrBadToken, "parseAdminAuth: checksum error")
		return
	}

	utype, err := acc.GetUtype(ctx, uid)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseAdminAuth: GetUtype - uid:", uid).Detail(err)
		return
	}

	user.Appid = appid
	user.Uid = uid
	// 防止su到比自己权限更高的用户上
	user.Utype = (utype &^ USER_TYPE_SUDOERS) | (utype & utypeSu & USER_TYPE_SUDOERS)
	user.UtypeSu = utypeSu
	user.Sudoer = info.Uid
	return
}

// --------------------------------------------------------------------
