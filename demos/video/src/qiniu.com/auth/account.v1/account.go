package account

import (
	"qbox.us/api/qconf/akg"
	"qbox.us/api/qconf/uidg"

	qconf "qbox.us/qconf/qconfapi"

	. "code.google.com/p/go.net/context"
	. "qiniu.com/auth/proto.v1"
)

// --------------------------------------------------------------------

type Config struct {
	Qconf         *qconf.Client `json:"-"`
	Qconfg        qconf.Config  `json:"qconfg"`
}

// --------------------------------------------------------------------

type Account struct {
	qconfg *qconf.Client
}

func New(cfg *Config) (r Account) {

	if cfg.Qconf == nil {
		r.qconfg = qconf.New(&cfg.Qconfg)
	} else {
		r.qconfg = cfg.Qconf
	}
	return
}

func (r Account) GetUtype(ctx Context, uid uint32) (utype uint32, err error) {

	return uidg.Client{r.qconfg}.GetUtype(nil, uid)
}

func (r Account) GetAccessInfo(ctx Context, accessKey string) (ret AccessInfo, err error) {

	info, err := akg.Client{r.qconfg}.Get(nil, accessKey)
	return AccessInfo(info), err
}

// --------------------------------------------------------------------

