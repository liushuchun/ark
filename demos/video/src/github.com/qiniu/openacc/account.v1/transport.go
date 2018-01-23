package account

import (
	"net/http"

	. "github.com/qiniu/openacc/account.api.v1"
)

const (
	authMethod = "Bearer "
)

// ---------------------------------------------------------------------------

type Transport struct {
	auth      string
	Transport http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	req.Header.Set("Authorization", t.auth)
	return t.Transport.RoundTrip(req)
}

func (t *Transport) NestedObject() interface{} {

	return t.Transport
}

func (p *Manager) NewTransport(user *UserInfo, transport http.RoundTripper) *Transport {

	if transport == nil {
		transport = http.DefaultTransport
	}
	auth := authMethod + p.MakeToken(user)
	return &Transport{auth, transport}
}

func (p *Manager) NewClient(user *UserInfo, transport http.RoundTripper) *http.Client {

	t := p.NewTransport(user, transport)
	return &http.Client{Transport: t}
}

// ---------------------------------------------------------------------------

