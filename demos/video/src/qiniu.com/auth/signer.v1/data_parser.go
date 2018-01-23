package signer

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strings"

	. "code.google.com/p/go.net/context"
	"github.com/qiniu/errors"

	"qiniu.com/auth/authutil.v1"
	. "qiniu.com/auth/proto.v1"
)

func signData(sk, data []byte) []byte {

	h := hmac.New(sha1.New, sk)
	h.Write(data)
	return h.Sum(nil)
}

// [:suinfo:]key:token:data
func ParseDataAuth(acc Interface, ctx Context,
	token string) (user SudoerInfo, data string, err error) {

	if strings.HasPrefix(token, ":") {
		return parseDataAdmin(acc, ctx, token)
	}
	return parseDataNormal(acc, ctx, token)
}

func parseDataNormal(acc Interface, ctx Context, token string) (user SudoerInfo, data string, err error) {

	pt := strings.SplitN(token, ":", 3)
	if len(pt) != 3 {
		err = ErrBadToken
		return
	}
	key, signExp, data := pt[0], pt[1], pt[2]

	info, err := acc.GetAccessInfo(ctx, key)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseDataNormal: GetAccessInfo").Detail(err)
		return
	}

	sign := signData(info.Secret, []byte(data))
	if base64.URLEncoding.EncodeToString(sign) != signExp {
		err = errors.Info(ErrBadToken, "parseDataNormal: checksum error")
		return
	}

	user.Access = key
	user.Appid = info.Appid
	user.Uid = info.Uid
	user.Utype, err = acc.GetUtype(ctx, user.Uid)
	return
}

func parseDataAdmin(acc Interface, ctx Context, token string) (user SudoerInfo, data string, err error) {

	pt := strings.SplitN(token, ":", 5)
	if len(pt) != 5 {
		err = ErrBadToken
		return
	}
	suInfo, key, signExp, data := pt[1], pt[2], pt[3], pt[4]
	uid, appid, err := authutil.ParseSuInfo(suInfo)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseSuInfo: ", suInfo).Detail(err)
		return
	}

	info, err := acc.GetAccessInfo(ctx, key)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseDataAdmin: GetAccessInfo").Detail(err)
		return
	}
	utypeSu, err := acc.GetUtype(ctx, info.Uid)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseDataAdmin: GetUtype").Detail(err)
		return
	}
	if (utypeSu & USER_TYPE_SUDOERS) == 0 {
		err = errors.Info(ErrBadToken, "parseDataAdmin: not sudoer")
		return
	}

	dataForSign := ":" + suInfo + ":" + data
	sign := signData(info.Secret, []byte(dataForSign))

	if base64.URLEncoding.EncodeToString(sign) != signExp {
		err = errors.Info(ErrBadToken, "parseDataAdmin: checksum error")
		return
	}

	utype, err := acc.GetUtype(ctx, uid)
	if err != nil {
		err = errors.Info(ErrBadToken, "parseDataAdmin: GetUtype - uid:", uid).Detail(err)
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
