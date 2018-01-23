package account

import (
	"strconv"
	"net/http"

	"github.com/qiniu/openacc/validation.v1/regexp"
	"qiniupkg.com/http/httputil.v2"

	. "github.com/qiniu/openacc/storage/proto.v1"
	. "github.com/qiniu/openacc/validation.v1"
)

// ---------------------------------------------------------------------------

type userName struct {
}

func UserName() userName {
	return userName{}
}

func (p userName) Validate(k, v string) error {
	for _, c := range v {
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
		if !ok {
			return httputil.NewError(400, "%s is invalid: must match [a-z0-9]+")
		}
	}
	return nil
}

// ---------------------------------------------------------------------------

type userNewRet struct {
	Id string `json:"id"`
}

/*
POST /v1/user/new
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer <AdminToken>

username=<UserName>&
password=<Password>&
email=<Email>&
email_status=<EmailStatus>&
utype=<Utype>

200 OK
Content-Type: application/json

{
  "id": <UserId>
}
*/
func (p *Service) PostUserNew(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		httputil.Error(w, ErrPostMethodOnly)
		return
	}

	admin, err := p.validateAdminToken(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	utype, err := strconv.ParseUint(req.Form.Get("utype"), 10, 0)
	if err != nil {
		httputil.Error(w, ErrInvalidUtype)
		return
	}

	err = p.canSetUtype(admin, int(utype))
	if err != nil {
		httputil.Error(w, err)
		return
	}

	name := req.Form.Get("username")
	if err := StringValidate("username", name, RangeLen(4, 20), UserName()); err != nil {
		httputil.Error(w, err)
		return
	}

	pwd := req.Form.Get("password")
	if err := StringValidate("password", pwd, MinLen(6)); err != nil {
		httputil.Error(w, err)
		return
	}

	email := req.Form.Get("email")
	if err := StringValidate("email", email, regexp.Email()); err != nil {
		httputil.Error(w, err)
		return
	}

	emailStatus, err := strconv.ParseUint(req.Form.Get("email_status"), 10, 0)
	if err != nil {
		httputil.Error(w, ErrInvalidEmailStatus)
		return
	}

	salt := genSalt()
	entry := &UserEntry{
		Name:        name,
		PwdMac:      calcPwdMac(pwd, salt),
		Salt:        salt,
		Email:       email,
		EmailStatus: int(emailStatus),
		Status:      1,
		Utype:       int(utype),
	}
	id, err := p.Storage.InsertUser(entry)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.Reply(w, 200, &userNewRet{id})
}

func genSalt() string {

	return ""
}

// ---------------------------------------------------------------------------

/*
POST /v1/user/update
Authorization: Bearer <AdminToken>

username=<UserName>& (或者email=<Email> 或者 uid=<UserId> 三选一）
password=<Password>&
new_email=<NewEmail>&
email_status=<EmailStatus>&
status=<Status>&
utype=<Utype>

200 OK
*/
func (p *Service) PostUserUpdate(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		httputil.Error(w, ErrPostMethodOnly)
		return
	}

	admin, err := p.validateAdminToken(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	entry := new(UserEntry)

	if val := req.Form.Get("utype"); val != "" {
		utype, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			httputil.Error(w, ErrInvalidUtype)
			return
		}
		if err := p.canSetUtype(admin, int(utype)); err != nil {
			httputil.Error(w, err)
			return
		}
		entry.Utype = int(utype)
	}

	if pwd := req.Form.Get("password"); pwd != "" {
		if err := StringValidate("password", pwd, MinLen(6)); err != nil {
			httputil.Error(w, err)
			return
		}
		salt := genSalt()
		entry.Salt = salt
		entry.PwdMac = calcPwdMac(pwd, salt)
	}

	if email := req.Form.Get("new_email"); email != "" {
		if err := StringValidate("email", email, regexp.Email()); err != nil {
			httputil.Error(w, err)
			return
		}
		entry.Email = email
	}

	if val := req.Form.Get("email_status"); val != "" {
		emailStatus, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			httputil.Error(w, ErrInvalidEmailStatus)
			return
		}
		entry.EmailStatus = int(emailStatus)
	}

	err = p.Storage.UpdateUser(userSelector(req), entry)
	httputil.Error(w, err)
}

// ---------------------------------------------------------------------------

type userInfoRet struct {
	Id          string `json:"id"`
	Name        string `json:"username"`
	Email       string `json:"email"`
	EmailStatus int    `json:"email_status"`
	Status      int    `json:"status"`
	Utype       int    `json:"utype"`
}

/*
GET /v1/user/info?username=<UserName> (或者email=<Email> 或者 uid=<UserId> 三选一）
Authorization: Bearer <AdminToken>

200 OK
Content-Type: application/json

{
  "id": <UserId>,
  "username": <UserName>,
  "email": <Email>,
  "email_status": <EmailStatus>,
  "status": <Status>,
  "utype": <Utype>
}
*/
func (p *Service) GetUserInfo(w http.ResponseWriter, req *http.Request) {

	_, err := p.validateAdminToken(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	id, entry, err := p.Storage.GetUser(userSelector(req))
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.Reply(w, 200, &userInfoRet{
		Id:          id,
		Name:        entry.Name,
		Email:       entry.Email,
		EmailStatus: entry.EmailStatus,
		Status:      entry.Status,
		Utype:       entry.Utype,
	})
}

// ---------------------------------------------------------------------------

func (p *Service) RegisterRouteAdmin(mux *http.ServeMux) {

	mux.HandleFunc("/v1/user/new", p.PostUserNew)
	mux.HandleFunc("/v1/user/update", p.PostUserUpdate)
	mux.HandleFunc("/v1/user/info", p.GetUserInfo)
}

// ---------------------------------------------------------------------------

