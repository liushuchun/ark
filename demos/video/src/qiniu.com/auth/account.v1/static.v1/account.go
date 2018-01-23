package static

import (
	"syscall"

	. "code.google.com/p/go.net/context"
	. "qiniu.com/auth/proto.v1"
)

// --------------------------------------------------------------------

type Info struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
	Uid    uint32 `json:"uid"`
	Utype  uint32 `json:"utype"`
}

type Config struct {
	Users []Info `json:"users"`
}

// --------------------------------------------------------------------

type Account struct {
	accessInfos map[string]AccessInfo // access -> Info
	utypes      map[uint32]uint32     // uid -> Info
}

func New(cfg *Config) (r Account) {

	accessInfos := make(map[string]AccessInfo, len(cfg.Users))
	utypes := make(map[uint32]uint32, len(cfg.Users))
	for _, u := range cfg.Users {
		accessInfos[u.Access] = AccessInfo{
			Secret: []byte(u.Secret),
			Uid:    u.Uid,
		}
		utypes[u.Uid] = u.Utype
	}
	r = Account{accessInfos: accessInfos, utypes: utypes}
	return
}

func (r Account) GetUtype(ctx Context, uid uint32) (utype uint32, err error) {
	utype, ok := r.utypes[uid]
	if !ok {
		err = syscall.ENOENT
		return
	}
	return
}

func (r Account) GetAccessInfo(ctx Context, accessKey string) (ret AccessInfo, err error) {
	ret, ok := r.accessInfos[accessKey]
	if !ok {
		err = syscall.ENOENT
		return
	}
	return
}

// --------------------------------------------------------------------
