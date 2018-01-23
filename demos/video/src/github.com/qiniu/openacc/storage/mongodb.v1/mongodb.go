package mongodb

import (
	"gopkg.in/mgo.v2"

	. "github.com/qiniu/openacc/storage/proto.v1"
)

// ---------------------------------------------------------------------------

type Impl struct {
	users *mgo.Collection
}

func New(session *mgo.Session) (p *Impl, err error) {

	return
}

func (p *Impl) InsertUser(entry *UserEntry) (id string, err error) {

	return
}

func (p *Impl) UpdateUser(sel *UserSelector, entry *UserEntry) (err error) {

	return
}

func (p *Impl) GetUser(sel *UserSelector) (id string, entry *UserEntry, err error) {

	return
}

// ---------------------------------------------------------------------------

