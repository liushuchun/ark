package mgoutil

import (
	"reflect"
	"syscall"
	"time"

	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

func Dail(host string, mode string, syncTimeoutInS int64) (session *mgo.Session, err error) {

	session, err = mgo.Dial(host)
	if err != nil {
		log.Error("Connect MongoDB failed:", err, "- host:", host)
		return
	}

	if mode != "" {
		SetMode(session, mode, true)
	}
	if syncTimeoutInS != 0 {
		session.SetSyncTimeout(time.Duration(int64(time.Second) * syncTimeoutInS))
	}
	return
}

// ------------------------------------------------------------------------

type Config struct {
	Host           string `json:"host"`
	DB             string `json:"db"`
	Mode           string `json:"mode"`
	SyncTimeoutInS int64  `json:"timeout"` // 以秒为单位
}

func Open(ret interface{}, cfg *Config) (session *mgo.Session, err error) {

	session, err = Dail(cfg.Host, cfg.Mode, cfg.SyncTimeoutInS)
	if err != nil {
		return
	}

	db := session.DB(cfg.DB)
	err = InitCollections(ret, db)
	if err != nil {
		session.Close()
		session = nil
	}
	return
}

func InitCollections(ret interface{}, db *mgo.Database) (err error) {

	v := reflect.ValueOf(ret)
	if v.Kind() != reflect.Ptr {
		log.Error("mgoutil.Open: ret must be a pointer")
		return syscall.EINVAL
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		log.Error("mgoutil.Open: ret must be a struct pointer")
		return syscall.EINVAL
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Tag == "" {
			continue
		}
		coll := sf.Tag.Get("coll")
		if coll == "" {
			continue
		}
		switch elem := v.Field(i).Addr().Interface().(type) {
		case *Collection:
			elem.Collection = db.C(coll)
		case **mgo.Collection:
			*elem = db.C(coll)
		default:
			log.Error("mgoutil.Open: coll must be *mgo.Collection or mgoutil.Collection")
			return syscall.EINVAL
		}
	}
	return
}

// ------------------------------------------------------------------------
