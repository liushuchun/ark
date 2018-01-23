package ta

import (
	"github.com/qiniu/reliable"
	. "github.com/qiniu/reliable/ta"
	"os"
)

type taTester interface {
	clear()
	setA(mid int)
	setB(mid int)
	checkA(mid int)
	checkB(mid int)
}

type testerCreater func(ta *Transaction) taTester

func testTaTester(loops int, creater testerCreater) {
	rlog := "test_rlog.qboxtest"
	defer os.Remove(rlog)

	var ta *Transaction
	var tester taTester
	inited := make([]bool, loops)

	clear := func() {
		if ta != nil {
			ta.Close()
			ta = nil
		}
		os.Remove(rlog)
		if tester != nil {
			tester.clear()
			tester = nil
		}
		inited = make([]bool, loops)
	}
	defer clear()

	create := func(reset bool) {
		if reset {
			clear()
		}

		file, err := reliable.OpenCfgfile([]string{rlog}, 1)
		if err != nil {
			panic(err)
		}
		ta = OpenTransaction(file)
		tester = creater(ta)
		ta.Setup()
	}

	const (
		A = "A"
		B = "B"
	)

	modify := func(mid int, ab string, commit bool, end bool, endfail bool) {
		invoke := func() {
			err := ta.Begin()
			if err != nil {
				panic(err)
			}
			if ab == A {
				tester.setA(mid)
				tester.checkA(mid)
			} else {
				tester.setB(mid)
				tester.checkB(mid)
			}
			if !commit {
				ta.Rollback()
			}
			if end {
				ta.EndWithFail(endfail)
			}
		}

		if inited[mid] {
			if ab == A {
				tester.checkB(mid)
			} else {
				tester.checkA(mid)
			}
		}

		invoke()

		if !commit {
			if inited[mid] {
				if ab == A {
					tester.checkB(mid)
				} else {
					tester.checkA(mid)
				}
			}
		} else {
			if ab == A {
				tester.checkA(mid)
			} else {
				tester.checkB(mid)
			}
		}

		if commit && end && !endfail {
			inited[mid] = true
		}
	}

	modifys := func(tester taTester, ab string, commit bool, end bool, endfail bool) {
		for i := 0; i < loops; i++ {
			modify(i, ab, commit, end, endfail)
		}
	}

	at := func() {
		modifys(tester, A, true, true, false)
	}
	bt := func() {
		modifys(tester, B, true, true, false)
	}
	af := func() {
		modifys(tester, A, false, true, false)
	}
	bf := func() {
		modifys(tester, B, false, true, false)
	}

	ac := func() {
		modifys(tester, A, false, false, false)
	}
	bc := func() {
		modifys(tester, B, false, false, false)
	}
	ad := func() {
		modifys(tester, A, true, false, false)
	}
	bd := func() {
		modifys(tester, B, true, false, false)
	}

	ax := func() {
		modifys(tester, A, false, true, true)
	}
	bx := func() {
		modifys(tester, B, false, true, true)
	}
	ay := func() {
		modifys(tester, A, true, true, true)
	}
	by := func() {
		modifys(tester, B, true, true, true)
	}

	confuns1 := []func(){at, bt, at, bf, bf, bf, bt, at, bt}
	intfuns1 := []func(){at, bf, bt, af, at, bf, bt, af, at}
	intfuns2 := []func(){af, at, bf, bt, af, at, bf, bt, af}

	rollback := func() {
		create(true)
		modify(0, A, false, true, false)
	}
	rollback()

	continuous := func() {
		create(true)
		for _, fun := range confuns1 {
			fun()
		}
	}
	continuous()

	interval1 := func() {
		create(true)
		for _, fun := range intfuns1 {
			fun()
		}
	}
	interval1()

	interval2 := func() {
		create(true)
		for _, fun := range intfuns2 {
			fun()
		}
	}
	interval2()

	reload := func(funs []func(), point int) {
		create(true)
		for i := 0; i <= point; i++ {
			funs[i]()
		}
		create(false)
		for i := point + 1; i < len(funs); i++ {
			funs[i]()
		}
	}

	for _, seq := range [][]func(){confuns1, intfuns1, intfuns2} {
		for i := 0; i < len(seq); i++ {
			reload(seq, i)
		}
	}

	loops = 1

	crash := func() {
		reload([]func(){ac, at, bf, bt}, 0)
		reload([]func(){ad, at, bf, bt}, 0)

		reload([]func(){at, bc, bt}, 1)
		reload([]func(){at, bd, bt}, 1)
		reload([]func(){at, bc, bf}, 1)
		reload([]func(){at, bd, bf}, 1)

		reload([]func(){at, bt, ac, at}, 2)
		reload([]func(){at, bt, ad, at}, 2)
		reload([]func(){at, bt, ac, at}, 2)
		reload([]func(){at, bt, ad, at}, 2)
	}
	crash()

	endfail := func() {
		reload([]func(){at, bx, bt}, 1)
		reload([]func(){at, by, bt}, 1)
		reload([]func(){at, bx, bf}, 1)
		reload([]func(){at, by, bf}, 1)

		reload([]func(){at, bt, ax, at}, 2)
		reload([]func(){at, bt, ay, at}, 2)
		reload([]func(){at, bt, ax, at}, 2)
		reload([]func(){at, bt, ay, at}, 2)
	}
	endfail()
}
