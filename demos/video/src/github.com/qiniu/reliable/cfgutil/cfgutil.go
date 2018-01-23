package cfgutil

import (
	"encoding/json"
	"github.com/qiniu/errors"
	"github.com/qiniu/reliable"
)

// ------------------------------------------------------------------------------------

func WriteFile(p *reliable.Config, v interface{}) (err error) {

	b, err := json.Marshal(v)
	if err != nil {
		err = errors.Info(err, "reliable/cfgutil.WriteFile: json.Marshal failed").Detail(err)
		return
	}

	err = p.WriteFile(b)
	if err != nil {
		err = errors.Info(err, "reliable/cfgutil.WriteFile failed").Detail(err)
	}
	return
}

func ReadFile(p *reliable.Config, v interface{}) (err error) {

	b, err := p.ReadFile()
	if err != nil {
		err = errors.Info(err, "reliable/cfgutil.ReadFile failed").Detail(err)
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		err = errors.Info(err, "reliable/cfgutil.ReadFile: json.Unmarshal failed").Detail(err)
	}
	return
}

// ------------------------------------------------------------------------------------
