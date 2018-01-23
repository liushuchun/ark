package configversion

import (
	"fmt"
	"sync"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DefaultReloadMs = 10000
)

type ReloadingConfig struct {
	Id       string `json:"conf_name"`
	ReloadMs int    `json:"reloading_ms"` // use DefaultReloadMs if zero.
	C        *mgo.Collection

	ver  uint32
	once sync.Once
}

func (cfg *ReloadingConfig) advance() error {
	err := cfg.C.UpdateId(cfg.Id, bson.M{"$inc": bson.M{"ver": 1}})
	return err
}

type entry struct {
	Id  string `bson:"_id"`
	Ver uint32 `bson:"ver"`
}

// StartReloading will call onReload every time when someone calls advance with same Id.
func StartReloading(
	cfg *ReloadingConfig, onReload func(l rpc.Logger) error) (advance func() error, err error) {

	cfg.once.Do(func() {
		err = startReloading(cfg, onReload)
	})
	if err == nil {
		advance = cfg.advance
	}
	return
}

func startReloading(
	cfg *ReloadingConfig, onReload func(l rpc.Logger) error) (err error) {

	xl := xlog.NewWith("StartReloading")

	sess := cfg.C.Database.Session.Copy()
	var shouldclose = true
	defer func() {
		if shouldclose {
			sess.Close()
		}
	}()
	coll := cfg.C.With(sess)

	var e entry
	err = coll.FindId(cfg.Id).One(&e)
	if err == mgo.ErrNotFound {
		err = coll.Insert(entry{Id: cfg.Id, Ver: 1})
		e.Ver = 1
		if mgo.IsDup(err) {
			err = coll.FindId(cfg.Id).One(&e)
		}
	}
	if err != nil {
		return
	}
	cfg.ver = e.Ver

	err = onReload(xl)
	if err != nil {
		xl.Error("reload:", errors.Detail(err))
		return
	}
	reloadMs := cfg.ReloadMs
	if reloadMs == 0 {
		reloadMs = DefaultReloadMs
	}
	shouldclose = false
	go func() {
		dur := time.Duration(reloadMs) * time.Millisecond
		defer func() {
			sess.Close()
		}()

		for t := range time.Tick(dur) {

			xl := xlog.NewWith(fmt.Sprintf("Reloading/%v/%v", cfg.Id, t.Unix()))
			err = coll.FindId(cfg.Id).One(&e)
			if err == mgo.ErrNotFound {
				continue
			}
			if err != nil {
				sess.Close()
				sess = cfg.C.Database.Session.Copy()
				coll = cfg.C.With(sess)
				xl.Error("Reloading: C.Findid", err)
				continue
			}
			if cfg.ver == e.Ver {
				continue
			}

			xl.Infof("ver is changed, confName: %v, oldVer: %v, newVer: %v", cfg.Id, cfg.ver, e.Ver)
			err := onReload(xl)
			if err != nil {
				xl.Error("Reloading: onReload:", errors.Detail(err))
				continue
			}
			cfg.ver = e.Ver
		}
	}()
	return
}
