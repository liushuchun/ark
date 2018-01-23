package account

import (
	"testing"

	. "github.com/qiniu/openacc/account.api.v1"
)

// ---------------------------------------------------------------------------

func TestToken(t *testing.T) {

	keyPairs := []KeyPair{
		{"ak1", "hello"},
		{"ak2", "xsw"},
	}

	cfg1 := &Config{
		KeyPairs: keyPairs[1:],
	}
	acc1, err := New(cfg1)
	if err != nil {
		t.Fatal("account.New failed:", err)
	}

	user1 := &UserInfo{
		Name: "test",
		Utype: UtypeAdmin,
		Expiry: 1,
	}
	token1 := acc1.MakeToken(user1)
	println("token:", token1)

	user1Parsed, err := acc1.ParseToken(token1)
	if err != nil {
		t.Fatal("acc1.ParseToken failed:", err)
	}
	if *user1 != *user1Parsed {
		t.Fatal("acc1.ParseToken failed: *user1 != *user1Parsed")
	}

	cfg2 := &Config{
		KeyPairs: keyPairs,
	}
	acc2, err := New(cfg2)
	if err != nil {
		t.Fatal("account.New failed:", err)
	}

	user2, err := acc2.ParseToken(token1)
	if err != nil {
		t.Fatal("acc2.ParseToken failed:", err)
	}
	if *user1 != *user2 {
		t.Fatal("acc2.ParseToken failed: *user1 != *user2")
	}
}

// ---------------------------------------------------------------------------

