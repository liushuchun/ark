package account

import (
	"crypto/md5"
	"encoding/base64"
	"net/http"
	"time"

	"qiniupkg.com/http/httputil.v2"

	. "github.com/qiniu/openacc/storage/proto.v1"
)

// ---------------------------------------------------------------------------

type userTokenRet struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in`
}

/*
POST /v1/user/token
Content-Type: application/x-www-form-urlencoded

grant_type=password&
client_id=<ClientAppId>&
device_id=<DeviceId>&
username=<UserName>& (或者email=<Email> 或者 uid=<UserId> 三选一）
password=<Password>&
scope=<Scope>

或：

POST /v1/user/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&
client_id=<ClientAppId>&
device_id=<DeviceId>&
refresh_token=<RefreshToken>&
scope=<Scope>

返回：

200 OK
Content-Type: application/json

{
  access_token: <AccessToken>
  token_type: <TokenType> #目前只有bearer
  expires_in: <ExpireSconds>
  refresh_token: <RefreshToken>
}
*/
func (p *Service) PostUserToken(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		httputil.Error(w, ErrPostMethodOnly)
		return
	}

	switch req.Form.Get("grant_type") {
	case "password":
		id, entry, err := p.Storage.GetUser(userSelector(req))
		if err != nil {
			httputil.Error(w, err)
			return
		}
		if calcPwdMac(req.Form.Get("password"), entry.Salt) != entry.PwdMac {
			httputil.Error(w, ErrBadUsernameOrPwd)
			return
		}
		expiry := time.Now().Unix() + 3600
		httputil.Reply(w, 200, &userTokenRet{
			AccessToken:  p.makeToken(id, entry, req, expiry),
			RefreshToken: "",
			TokenType:    "bearer",
			ExpiresIn:    expiry,
		})
	case "refresh_token":
		httputil.Error(w, ErrNotImpl)
	default:
		httputil.Error(w, ErrNotImpl)
	}
}

func userSelector(req *http.Request) *UserSelector {

	return &UserSelector{
		Id: req.Form.Get("uid"),
		Name: req.Form.Get("username"),
		Email: req.Form.Get("email"),
	}
}

func calcPwdMac(pwd, salt string) string {

	b := make([]byte, len(pwd) + len(salt))
	copy(b, pwd)
	copy(b[len(pwd):], salt)
	hash := md5.Sum(b)
	mac := base64.URLEncoding.EncodeToString(hash[:])
	return mac[:22]
}

// ---------------------------------------------------------------------------

/*
POST /v1/user/logout
Content-Type: application/x-www-form-urlencoded

refresh_token=<RefreshToken>

200 OK
*/
func (p *Service) PostUserLogout(w http.ResponseWriter, req *http.Request) {

}

// ---------------------------------------------------------------------------

func (p *Service) RegisterRoute(mux *http.ServeMux) {

	mux.HandleFunc("/v1/user/token", p.PostUserToken)
	mux.HandleFunc("/v1/user/logout", p.PostUserLogout)
}

// ---------------------------------------------------------------------------

