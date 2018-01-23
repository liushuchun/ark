package ta

import (
	. "github.com/qiniu/reliable/ta"
	"os"
	"testing"
)

type cpxTester struct {
	creaters []testerCreater
	testers  []taTester
}

func createCpxTester(creaters []testerCreater, ta *Transaction) taTester {
	t := &cpxTester{creaters, make([]taTester, len(creaters))}
	for i, creater := range t.creaters {
		t.testers[i] = creater(ta)
	}
	return t
}

func (t *cpxTester) clear() {
	for _, tester := range t.testers {
		tester.clear()
	}
}

func (t *cpxTester) setA(mid int) {
	for _, tester := range t.testers {
		tester.setA(mid)
	}
}

func (t *cpxTester) setB(mid int) {
	for _, tester := range t.testers {
		tester.setB(mid)
	}
}

func (t *cpxTester) checkA(mid int) {
	for _, tester := range t.testers {
		tester.checkA(mid)
	}
}

func (t *cpxTester) checkB(mid int) {
	for _, tester := range t.testers {
		tester.checkB(mid)
	}
}

func TestComplexTa(t *testing.T) {
	f0 := "test_config.qboxtest"
	f1 := "test_array.qboxtest"
	f2 := "test_bigarray.qboxtest"
	f3 := "test_bitmap.qboxtest"
	f4 := "test_bigbitmap.qboxtest"
	defer os.Remove(f0)
	defer os.Remove(f1)
	defer os.Remove(f2)
	defer os.Remove(f3)
	defer os.Remove(f4)

	creaters := []testerCreater{
		func(ta *Transaction) taTester {
			return createConfigTester(f0, 0, ta)
		},
		func(ta *Transaction) taTester {
			return createArrayTester(f1, 1, ta)
		},
		func(ta *Transaction) taTester {
			return createBigArrayTester(f2, 2, ta)
		},
		func(ta *Transaction) taTester {
			return createBitmapTester(f3, 3, ta)
		},
		func(ta *Transaction) taTester {
			return createBigBitmapTester(f4, 4, ta)
		},
	}
	testTaTester(17, func(ta *Transaction) taTester {
		return createCpxTester(creaters, ta)
	})
}
