package mgoutil

import (
	"reflect"
	"strings"
	"testing"

	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

func doTestParseIndex(t *testing.T, colIndex string, expected mgo.Index) {

	var index mgo.Index

	pos := strings.Index(colIndex, ":")
	if pos >= 0 {
		parseIndexOptions(&index, colIndex[pos+1:])
		colIndex = colIndex[:pos]
	}
	index.Key = strings.Split(strings.TrimRight(colIndex, " "), ",")

	if !reflect.DeepEqual(index, expected) {
		t.Fatal("parseIndex failed:", colIndex, "expected:", expected, "real:", index)
	}
}

func TestParseIndex(t *testing.T) {

	doTestParseIndex(
		t, "uid,status,delete :unique,sparse",
		mgo.Index{Key: []string{"uid", "status", "delete"}, Sparse: true, Unique: true})

	doTestParseIndex(
		t, "email :background",
		mgo.Index{Key: []string{"email"}, Background: true})
}

// ------------------------------------------------------------------------
