package table_test

import (
	"fmt"
	"testing"

	"github.com/qiniu/ds/mgotable.v1"
	"github.com/qiniu/ds/table.v1"
	"gopkg.in/mgo.v2"
)

// ---------------------------------------------------------------------------

type M map[string]interface{}

func TestMgoTable(t *testing.T) {

	fmt.Println("------------ TestMgoTable -----------")

	session, err := mgo.Dial("localhost")
	if err != nil {
		t.Fatal("mgo.Dial failed:", err)
	}
	defer session.Close()

	c := session.DB("table_proto").C("test")
	c.RemoveAll(M{})

	creator := mgotable.NewCreator(c, new(Row))
	doTestTable(t, creator)
}

func TestNormalTable(t *testing.T) {

	fmt.Println("------------ TestNormalTable -----------")

	creator := table.NewCreator(new(Row))
	doTestTable(t, creator)
}

// ---------------------------------------------------------------------------

