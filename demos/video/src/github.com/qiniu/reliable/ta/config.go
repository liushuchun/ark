package ta

import (
	"github.com/qiniu/errors"
	"github.com/qiniu/reliable"
	. "github.com/qiniu/reliable/errors"
	"syscall"
)

// --------------------------------------------------------------------
// type Config

type Config struct {
	*reliable.Config
	data []byte
	ta   *Transaction
	id   int
}

func OpenConfig(cfg *reliable.Config, ta *Transaction, id int) (p *Config, err error) {

	data, err := cfg.ReadFile()
	if err != nil {
		if cfg.Validate() != syscall.ENOENT {
			err = errors.Info(err, "ta.OpenConfig failed", id).Detail(err)
			return
		}
	}

	p = &Config{cfg, data, ta, id}
	err = ta.init(id, p)
	if err != nil {
		err = errors.Info(err, "ta.OpenConfig failed", id).Detail(err)
	}
	return
}

func (p *Config) WriteFile(data []byte) (err error) {

	rl, hint, err := p.ta.beginRlog(p.id)
	if err != nil {
		err = errors.Info(err, "ta.Config.WriteFile: beginRlog failed").Detail(err)
		return
	}
	rl.Write(p.data)
	rl.end(hint)

	err = p.Config.WriteFile(data)
	if err != nil {
		err = errors.Info(err, "ta.Config.WriteFile failed").Detail(err)
		return
	}
	p.data = data
	return
}

func (p *Config) ReadFile() ([]byte, error) {

	if len(p.data) == 0 {
		return nil, ErrBadData
	}
	return p.data, nil
}

func (p *Config) Validate() (err error) {

	if len(p.data) == 0 {
		return syscall.ENOENT
	}
	return nil
}

func (p *Config) DoAct(act []byte) (err error) {

	err = p.Config.WriteFile(act)
	if err != nil {
		err = errors.Info(err, "ta.Config.DoAct failed").Detail(err)
		return
	}

	data := make([]byte, len(act))
	copy(data, act)
	p.data = data
	return nil
}

// --------------------------------------------------------------------
