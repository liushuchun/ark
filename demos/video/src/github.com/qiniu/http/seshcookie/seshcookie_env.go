package seshcookie

import (
	"errors"
	"github.com/qiniu/http/misc/seshcookie"
	"net/http"
)

var ErrNoManager = errors.New("session manager not found")

// ---------------------------------------------------------------------------

type Getter interface {
	GetSessions(req *http.Request) *seshcookie.SessionManager
}

// ---------------------------------------------------------------------------

type Manager struct {
	sessions *seshcookie.SessionManager
}

func (em *Manager) InitSessions(cookieName, cookieDomain, key string) {
	em.sessions = seshcookie.NewSessionManager(cookieName, cookieDomain, key)
}

func (em Manager) GetSessions(req *http.Request) *seshcookie.SessionManager {
	return em.sessions
}

// ---------------------------------------------------------------------------

type Env struct {
	W       http.ResponseWriter
	Req     *http.Request
	Session map[string]interface{}
}

func (p *Env) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) error {

	if g, ok := rcvr.(Getter); ok {
		if sm := g.GetSessions(req); sm != nil {
			sw, session := sm.Get(*w, req)
			*w, p.W, p.Req, p.Session = sw, sw, req, session
			return nil
		}
	}
	return ErrNoManager
}

func (p *Env) CloseEnv() {
}

// ---------------------------------------------------------------------------
