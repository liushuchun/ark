package proto

import (
	"github.com/qiniu/openacc/account.api.v1"
)

// ---------------------------------------------------------------------------

const (
	EmailVerified   = 1
	EmailUnverified = 0x80
)

const (
	StatusNormal   = 1
	StatusDisabled = 0x80
)

const (
	UtypeAdmin    = account.UtypeAdmin
	UtypeUser     = account.UtypeUser
	UtypeExpUser  = account.UtypeExpUser
	UtypeDisabled = account.UtypeDisabled
)

type UserEntry struct {
	Name        string
	PwdMac      string
	Salt        string
	Email       string
	EmailStatus int
	Status      int
	Utype       int
}

type UserSelector struct {
	Id    string
	Name  string
	Email string
}

type Storage interface {
	InsertUser(entry *UserEntry) (id string, err error)
	UpdateUser(sel *UserSelector, entry *UserEntry) (err error)
	GetUser(sel *UserSelector) (id string, entry *UserEntry, err error)
}

// ---------------------------------------------------------------------------

