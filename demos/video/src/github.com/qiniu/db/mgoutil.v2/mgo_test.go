package mgoutil

import (
	"testing"

	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

type testConfig struct {
	A Collection      `coll:"a"`
	B *mgo.Collection `coll:"b"`
}

func TestMgo(t *testing.T) {

	var ret testConfig
	session, err := Open(&ret, &Config{Host: "localhost", DB: "test_mgoutil"})
	if err != nil {
		t.Fatal("Open failed:", err)
	}
	defer session.Close()

	if ret.A.Name != "a" {
		t.Fatal(`ret.A.Name != "a"`)
	}
	if ret.B.Name != "b" {
		t.Fatal(`ret.B.Name != "b"`)
	}
}

// ------------------------------------------------------------------------
