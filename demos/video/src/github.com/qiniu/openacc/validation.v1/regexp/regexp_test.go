package regexp

import (
	"testing"

	. "github.com/qiniu/openacc/validation.v1"
)

// ---------------------------------------------------------------------------

func Test(t *testing.T) {

	err := StringValidate("email", "test@qiniu.com", Email())
	if err != nil {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "test#qiniu.com", Email())
	if err == nil || err.Error() != "email is invalid: must be a valid email address" {
		t.Fatal("StringValidate failed:", err)
	}
}

// ---------------------------------------------------------------------------

