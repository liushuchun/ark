package account

import (
	"time"
)

// ---------------------------------------------------------------------------

const (
	UtypeAdmin      = 0x0001 // 管理员，可创建/修改非管理员用户
	UtypeUser       = 0x0004 // 普通用户
	UtypeExpUser    = 0x0010 // 体验用户
	UtypeDisabled   = 0x8000 // 被禁止的用户
)

type UserInfo struct {
	Id     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Appid  string `json:"app,omitempty"`
	Devid  string `json:"dev,omitempty"`
	Expiry int64  `json:"expiry"`
	Utype  int    `json:"role"`
}

func (p *UserInfo) IsValid() bool {

	return (p.Expiry + 7) >= time.Now().Unix()
}

func (p *UserInfo) IsAdmin() bool {

	return (p.Utype & UtypeAdmin) != 0
}

func (p *UserInfo) IsValidAdmin() bool {

	return p.IsValid() && p.IsAdmin()
}

// ---------------------------------------------------------------------------

