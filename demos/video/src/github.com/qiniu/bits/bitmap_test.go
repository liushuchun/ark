package bits

import (
	bits "github.com/qiniu/bits/testing"
	"testing"
)

// -----------------------------------------------------------

func TestBitmap(t *testing.T) {

	b := NewBitmap(nil)
	bits.BitmapTest(b, t)
}

// -----------------------------------------------------------
