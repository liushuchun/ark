package table

import (
	tablecfg "github.com/qiniu/ds/table.config.v1"
	"github.com/qiniu/ds/table.proto.v1"
)

// ---------------------------------------------------------------------------

type Creator struct {
	Cfg *tablecfg.Config
}

func NewCreator(elem interface{}) table.Creator {

	cfg := NewConfig(elem)
	return Creator{cfg}
}

func (p Creator) WithUniques(uniques ...interface{}) table.Creator {

	p.Cfg.WithUniques(uniques...)
	return p
}

func (p Creator) WithIndexes(indexes ...interface{}) table.Creator {

	p.Cfg.WithIndexes(indexes...)
	return p
}

func (p Creator) New() (table.Table, error) {

	return New(p.Cfg), nil
}

// ---------------------------------------------------------------------------

func (p *Table) CopySession() table.Table {

	return p
}

func (p *Table) CloseSession() error {

	return nil
}

// ---------------------------------------------------------------------------

