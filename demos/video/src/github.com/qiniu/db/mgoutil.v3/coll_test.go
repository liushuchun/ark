package mgoutil

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/mgo.v2"
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

func doTestParseIndexByType(t *testing.T, options string, expected mgo.Index) {

	var index mgo.Index

	parseIndexOptionsByType(&index, options)
	if !reflect.DeepEqual(index, expected) {
		t.Fatal("parseIndex failed:", options, "expected:", expected, "real:", index)
	}
}

func TestParseIndexByType(t *testing.T) {

	doTestParseIndexByType(
		t, "+name+time,unique",
		mgo.Index{Key: []string{"name", "time"}, Unique: true})

	doTestParseIndexByType(
		t, "unique", mgo.Index{Unique: true})

	doTestParseIndexByType(
		t, "index", mgo.Index{})
}

// ------------------------------------------------------------------------

