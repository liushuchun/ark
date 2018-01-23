package auth

import (
	. "code.google.com/p/go.net/context"
	"net/http"
	"strings"

	qaccount "qbox.us/account"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
	. "qiniu.com/auth/proto.v1"
	"qiniu.com/auth/qboxmac.v1"
	"qiniu.com/auth/qiniumac.v1"
	"qiniu.com/auth/signer.v1"
)

var (
	g_acc qaccount.Account
)

type AuthParser struct {
	AllowBearer bool
	AllowQBox   bool
	AllowQiniu  bool
	Acc         Interface
}

func (ap *AuthParser) ParseAuth(req *http.Request) (user account.UserInfo, err error) {

	authl, ok := req.Header["Authorization"]
	if !ok {
		err = signer.ErrBadToken
		return
	}

	var user1 SudoerInfo
	auth := authl[0]
	switch {
	case ap.AllowBearer && strings.HasPrefix(auth, "Bearer "):
		token := auth[7:]
		old, err1 := g_acc.ParseAccessToken(token)
		err = err1
		user1.Uid = old.Uid
		user1.Utype = old.Utype
		user1.Appid = uint64(old.Appid)
	case ap.AllowQBox && strings.HasPrefix(auth, "QBox "):
		token := auth[5:]
		user1, err = signer.ParseNormalAuth(qboxmac.DefaultRequestSigner, ap.Acc, Background(), token, req)
	case ap.AllowQBox && strings.HasPrefix(auth, "QBoxAdmin "):
		token := auth[10:]
		user1, err = signer.ParseAdminAuth(qboxmac.DefaultRequestSigner, ap.Acc, Background(), token, req)
	case ap.AllowQiniu && strings.HasPrefix(auth, "Qiniu "):
		token := auth[6:]
		user1, err = signer.ParseNormalAuth(qiniumac.DefaultRequestSigner, ap.Acc, Background(), token, req)
	case ap.AllowQiniu && strings.HasPrefix(auth, "QiniuAdmin "):
		token := auth[11:]
		user1, err = signer.ParseAdminAuth(qiniumac.DefaultRequestSigner, ap.Acc, Background(), token, req)
	default:
		err = signer.ErrBadToken
		return
	}

	user.Uid = user1.Uid
	user.Sudoer = user1.Sudoer
	user.Utype = user1.Utype
	user.UtypeSu = user1.UtypeSu
	user.Appid = uint32(user1.Appid)

	return
}

func (ap *AuthParser) ParseReverseTransport(req *http.Request, tr http.RoundTripper) (transport http.RoundTripper, err error) {

	authl, ok := req.Header["Authorization"]
	if !ok {
		err = signer.ErrBadToken
		return
	}

	auth := authl[0]
	var user1 SudoerInfo
	switch {
	case ap.AllowBearer && strings.HasPrefix(auth, "Bearer "):
		token := auth[7:]
		_, err1 := g_acc.ParseAccessToken(token)
		err = err1
		if err != nil {
			return
		}
		transport = oauth.NewTransport(token, tr)
	case ap.AllowQBox && strings.HasPrefix(auth, "QBox "):
		token := auth[5:]
		user1, err = signer.ParseNormalAuth(qboxmac.DefaultRequestSigner, ap.Acc, Background(), token, req)
		if err != nil {
			return
		}
		ak := user1.Access
		ret, err1 := ap.Acc.GetAccessInfo(Background(), ak)
		if err1 != nil {
			err = err1
			return
		}
		sk := ret.Secret
		mac := &qboxmac.Mac{ak, sk}
		transport = qboxmac.NewTransport(mac, tr)
	case ap.AllowQBox && strings.HasPrefix(auth, "QBoxAdmin "):
		token := auth[10:]
		user1, err = signer.ParseAdminAuth(qboxmac.DefaultRequestSigner, ap.Acc, Background(), token, req)
		if err != nil {
			return
		}
		ak := user1.Access
		ret, err1 := ap.Acc.GetAccessInfo(Background(), ak)
		if err1 != nil {
			err = err1
			return
		}
		sk := ret.Secret
		mac := &qboxmac.Mac{ak, sk}
		pos := strings.Index(token, ":")
		if pos == -1 {
			err = signer.ErrBadToken
			return
		}
		suInfo := token[:pos]
		transport = qboxmac.NewAdminTransport(mac, suInfo, tr)
	case ap.AllowQiniu && strings.HasPrefix(auth, "Qiniu "):
		token := auth[6:]
		user1, err = signer.ParseNormalAuth(qiniumac.DefaultRequestSigner, ap.Acc, Background(), token, req)
		if err != nil {
			return
		}
		ak := user1.Access
		ret, err1 := ap.Acc.GetAccessInfo(Background(), ak)
		if err1 != nil {
			err = err1
			return
		}
		sk := ret.Secret
		mac := &qiniumac.Mac{ak, sk}
		transport = qiniumac.NewTransport(mac, tr)
	case ap.AllowQiniu && strings.HasPrefix(auth, "QiniuAdmin "):
		token := auth[11:]
		user1, err = signer.ParseAdminAuth(qiniumac.DefaultRequestSigner, ap.Acc, Background(), token, req)
		if err != nil {
			return
		}
		ak := user1.Access
		ret, err1 := ap.Acc.GetAccessInfo(Background(), ak)
		if err1 != nil {
			err = err1
			return
		}
		sk := ret.Secret
		mac := &qiniumac.Mac{ak, sk}
		pos := strings.Index(token, ":")
		if pos == -1 {
			err = signer.ErrBadToken
			return
		}
		suInfo := token[:pos]
		transport = qiniumac.NewAdminTransport(mac, suInfo, tr)
	default:
		err = signer.ErrBadToken
		return
	}
	return
}
