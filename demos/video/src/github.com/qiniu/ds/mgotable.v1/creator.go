package mgotable

import (
	tablecfg "github.com/qiniu/ds/table.config.v1"
	"github.com/qiniu/ds/table.proto.v1"
	"gopkg.in/mgo.v2"
)

// ---------------------------------------------------------------------------

type Creator struct {
	coll *mgo.Collection
	cfg  *tablecfg.Config
}

func NewCreator(c *mgo.Collection, elem interface{}) table.Creator {

	cfg := NewConfig(elem)
	return &Creator{c, cfg}
}

func (p *Creator) WithUniques(uniques ...interface{}) table.Creator {

	p.cfg.WithUniques(uniques...)
	return p
}

func (p *Creator) WithIndexes(indexes ...interface{}) table.Creator {

	p.cfg.WithIndexes(indexes...)
	return p
}

func (p *Creator) New() (table.Table, error) {

	return New(p.coll, p.cfg)
}

// ---------------------------------------------------------------------------

func (p Table) CopySession() table.Table {

	c := p.Coll
	db := c.Database
	copy := db.Session.Copy().DB(db.Name).C(c.Name)
	return Table{copy}
}

func (p Table) CloseSession() error {

	p.Coll.Database.Session.Close()
	return nil
}

// ---------------------------------------------------------------------------

