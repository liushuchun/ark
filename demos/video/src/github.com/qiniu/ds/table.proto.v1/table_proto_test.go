package table_test

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/qiniu/ds/table.proto.v1"
)

// ---------------------------------------------------------------------------

type Row struct {
	Id   int     `bson:"id"`
	Pid  int     `bson:"pid"`
	Name string  `bson:"name"`
	File string  `bson:"file"`
}

type ById struct {
	Id int `bson:"id"`
}

type ByPid struct {
	Pid int `bson:"pid"`
}

type ByPidName struct {
	Pid  int    `bson:"pid"`
	Name string `bson:"name"`
}

func doTestTable(t *testing.T, creator table.Creator) {

	coll, err := creator.
		WithUniques(new(ById), new(ByPidName)).
		WithIndexes(new(ByPid)).
		New()

	if err != nil {
		t.Fatal("table.New failed:", err)
	}

	err = coll.Insert(
		&Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"},
		Row{Id: 2, Pid: 0, Name: "ttt"},
		&Row{Id: 3, Pid: 2, Name: "1.txt", File: "qiniu"},
	)
	if err != nil {
		t.Fatal("Insert failed:", err)
	}

	var rows1 []Row
	err = coll.FindAll(&rows1, ByPid{0})
	if err != nil {
		t.Fatal("Find by pid=0 failed:", err)
	}
	fmt.Println("rows1:", rows1)
	if len(rows1) != 2 {
		t.Fatal("Find by pid=0 failed - rows1")
	}

	var rows11 []*Row
	err = coll.FindAll(&rows11, ByPid{0})
	if err != nil {
		t.Fatal("Find by pid=0 failed:", err)
	}
	fmt.Println("rows11:", rows11)

	var rows2 *Row
	err = coll.FindOne(&rows2, ById{2})
	if err != nil {
		t.Fatal("Find by id=2 failed:", err)
	}
	fmt.Println("rows2:", *rows2)
	if (*rows2 != Row{Id: 2, Pid: 0, Name: "ttt"}) {
		t.Fatal("Find by id=2 failed - row2")
	}

	var rows3 Row
	err = coll.FindOne(&rows3, &ByPidName{0, "abc.doc"})
	if err != nil {
		t.Fatal("Find by pid=0, name=abc.doc failed:", err)
	}
	fmt.Println("rows3:", rows3)
	if (rows3 != Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"}) {
		t.Fatal("Find by pid=0, name=abc.doc failed - rows3")
	}

	coll.RemoveAll(ById{2})

	var rows4 []Row
	err = coll.FindAll(&rows4, ByPid{0})
	if err != nil {
		t.Fatal("Find by pid=0 failed:", err)
	}
	fmt.Println("rows4:", rows4)
	if len(rows4) != 1 || (rows4[0] != Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"}) {
		t.Fatal("Find by pid=0 failed - rows4")
	}

	coll.RemoveAll(ByPid{0})

	var rows5 *Row
	err = coll.FindOne(&rows5, ById{1})
	if err != syscall.ENOENT {
		t.Fatal("Find by id=1 failed - rows5 -", err)
	}

	err = coll.Insert(
		&Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"},
		&Row{Id: 2, Pid: 0, Name: "ttt"},
		&Row{Id: 3, Pid: 2, Name: "1.txt", File: "qiniu"},
	)
	if err != syscall.EEXIST {
		t.Fatal("Insert dup?")
	}

	var rows6 []Row
	err = coll.FindAll(&rows6, ByPid{0})
	if err != nil {
		t.Fatal("Find by pid=0 failed:", err)
	}
	fmt.Println("rows6:", rows6)
	if len(rows6) != 2 {
		t.Fatal("Find by pid=0 failed - rows6")
	}

	coll.RemoveAll(ByPid{0})

	var rows7 *Row
	err = coll.FindOne(&rows7, ById{1})
	if err != syscall.ENOENT {
		t.Fatal("Find by id=1 failed - rows7 -", err)
	}
}

// ---------------------------------------------------------------------------

