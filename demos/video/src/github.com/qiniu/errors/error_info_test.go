package errors

import (
	"errors"
	"syscall"
	"testing"
)

func TestCmd(t *testing.T) {

	err := Info(New("xxxx"), "foo.Bar failed: abc")
	cmd, ok := err.Method()
	if !ok || cmd != "foo.Bar" {
		t.Fatal("Invalid err.Method:", cmd)
	}
	msg := err.LogMessage()
	if msg != `foo.Bar failed:
 ==> github.com/qiniu/errors/error_info_test.go:11: xxxx ~ foo.Bar failed: abc` {
		t.Fatal("Invalid err.LogMessage:", msg)
	}
}

func MysqlError(err error, cmd ...interface{}) error {

	return InfoEx(2, syscall.EINVAL, cmd...).Detail(err)
}

func TestErrorsInfo(t *testing.T) {

	err := errors.New("detail error")
	err = MysqlError(err, "TestErrorsInfo failed")
	msg := Detail(err)
	if msg != ` ==> github.com/qiniu/errors/error_info_test.go:31: invalid argument ~ TestErrorsInfo failed
 ==> detail error` {
		t.Fatal("TestErrorsInfo failed")
	}
}

