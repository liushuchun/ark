package ta

import (
	"github.com/qiniu/reliable"
	. "github.com/qiniu/reliable/ta"
	"os"
	"testing"
)

type configTester struct {
	config *Config
	fname  string
}

func createConfigTester(fname string, id int, ta *Transaction) taTester {
	config1, err := reliable.OpenCfgfile([]string{fname}, 1)
	if err != nil {
		panic(err)
	}
	config2, err := OpenConfig(config1, ta, id)
	if err != nil {
		panic(err)
	}
	return &configTester{config2, fname}
}

func (t *configTester) clear() {
	os.Remove(t.fname)
}

func (t *configTester) set(mid int, val string) {
	if mid != 0 {
		return
	}
	err := t.config.WriteFile([]byte(val))
	if err != nil {
		panic(err)
	}
}

func (t *configTester) check(mid int, val string) {
	if mid != 0 {
		return
	}
	data, err := t.config.ReadFile()
	if err != nil {
		panic(err)
	}
	if string(data) != val {
		panic("got: " + string(data) + ", should be: " + val)
	}
}

func (t *configTester) setA(mid int) {
	t.set(mid, "abcdefghi")
}

func (t *configTester) setB(mid int) {
	t.set(mid, "ABCDEFGHIJKLMN")
}

func (t *configTester) checkA(mid int) {
	t.check(mid, "abcdefghi")
}

func (t *configTester) checkB(mid int) {
	t.check(mid, "ABCDEFGHIJKLMN")
}

func TestConfig(t *testing.T) {
	fname := "test_config.qboxtest"
	defer os.Remove(fname)
	testTaTester(10, func(ta *Transaction) taTester {
		return createConfigTester(fname, 0, ta)
	})
}
