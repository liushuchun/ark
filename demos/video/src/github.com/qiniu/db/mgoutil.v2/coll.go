package mgoutil

import (
	"strings"

	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

type Collection struct {
	*mgo.Collection
}

// ensure indexes for a collection
//
// eg. c.EnsureIndexes(
//			"uid :unique", "email :unique",
//			"serial_num", "uid,status,delete :sparse,background")
//
func (c Collection) EnsureIndexes(indexes ...string) {

	for _, colIndex := range indexes {
		var index mgo.Index
		pos := strings.Index(colIndex, ":")
		if pos >= 0 {
			parseIndexOptions(&index, colIndex[pos+1:])
			colIndex = colIndex[:pos]
		}
		index.Key = strings.Split(strings.TrimRight(colIndex, " "), ",")
		err := c.EnsureIndex(index)
		if err != nil {
			log.Fatal("<Mongo.C>:", c.Name, "Index:", index.Key, " error:", err)
			break
		}
	}
}

func parseIndexOptions(index *mgo.Index, options string) {

	for {
		var option string
		pos := strings.Index(options, ",")
		if pos < 0 {
			option = options
		} else {
			option = options[:pos]
			options = options[pos+1:]
		}
		switch option {
		case "unique":
			index.Unique = true
		case "sparse":
			index.Sparse = true
		case "background":
			index.Background = true
		default:
			log.Fatal("Unknown option:", option)
		}
		if pos < 0 {
			return
		}
	}
}

func (c Collection) CopySession() Collection {

	db := c.Database
	return Collection{db.Session.Copy().DB(db.Name).C(c.Name)}
}

func (c Collection) CloseSession() (err error) {

	c.Database.Session.Close()
	return nil
}

// ------------------------------------------------------------------------
