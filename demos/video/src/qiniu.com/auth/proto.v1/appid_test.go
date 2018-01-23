package proto

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

// --------------------------------------------------------------------

type OldAccessInfo struct {
	Secret []byte `bson:"secret"`
	Uid    uint32 `bson:"uid"`            // UserId
	AppId  uint32 `bson:"appId,omitempty"`
}

func TestAppid(t *testing.T) {

	old := OldAccessInfo{Secret: []byte("hello"), Uid: 1, AppId: 5}
	oldmsg, err := bson.Marshal(old)
	if err != nil {
		t.Fatal("bson.Marshal failed:", err)
	}

	var new AccessInfo
	err = bson.Unmarshal(oldmsg, &new)
	if err != nil {
		t.Fatal("bson.Unmarshal failed:", err)
	}
	fmt.Println("AccessInfo:", new)

	if string(new.Secret) != "hello" || new.Uid != 1 || new.Appid != 5 {
		t.Fatal(`new.Secret != "hello" || new.Uid != 1 || new.Appid != 5`)
	}

	newmsg, err := bson.Marshal(&new)
	if err != nil {
		t.Fatal("bson.Marshal(new) failed:", err)
	}

	var old2 OldAccessInfo
	err = bson.Unmarshal(newmsg, &old2)
	if err != nil {
		t.Fatal("bson.Unmarshal(newmsg) failed:", err)
	}
	if !reflect.DeepEqual(old, old2) {
		t.Fatal("old != old2")
	}
}

// --------------------------------------------------------------------

