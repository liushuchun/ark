package uuid

import (
	"testing"
)

// ---------------------------------------------------------------------------

func Test(t *testing.T) {

	m := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s, err := Gen(12)
		if err != nil {
			t.Fatal("Genguid failed:", err)
		}
		if m[s] {
			t.Fatal("Genguid conflict")
		}
		m[s] = true
		println(s)
	}
}

// ---------------------------------------------------------------------------
