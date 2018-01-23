package bits

import (
	bits "github.com/qiniu/bits/testing"
	"testing"
)

// -----------------------------------------------------------

func TestBigBitmap(t *testing.T) {

	b := NewBigBitmap(nil, 1<<25)
	bits.BitmapTest(b, t)
}

// -----------------------------------------------------------
