package validation

import (
	"testing"
)

// ---------------------------------------------------------------------------

func Test(t *testing.T) {

	err := StringValidate("email", "123", Required())
	if err != nil {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "", Required())
	if err == nil || err.Error() != "email is required" {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "1234", RangeLen(4, 10))
	if err != nil {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "123", RangeLen(4, 10))
	if err == nil || err.Error() != "email is too short: minimum length is 4" {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "123", MinLen(4))
	if err == nil || err.Error() != "email is too short: minimum length is 4" {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "123", MaxLen(3))
	if err != nil {
		t.Fatal("StringValidate failed:", err)
	}

	err = StringValidate("email", "123", MaxLen(2))
	if err == nil || err.Error() != "email is too long: maximum length is 2" {
		t.Fatal("StringValidate failed:", err)
	}
}

// ---------------------------------------------------------------------------

