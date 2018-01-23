package utf8

import (
	"testing"
)

type T struct {
	Intf interface{}
	Is   bool
}

func Test(t *testing.T) {

	validStr := "hello 世界"
	invalidStr := string([]byte{0xff, 0xfe, 0xfd})
	invalidStr2 := string([]byte{0xef, 0xbf, 0xbd})

	cases := []T{
		{
			Intf: validStr,
			Is:   true,
		},
		{
			Intf: invalidStr,
			Is:   false,
		},
		{
			Intf: invalidStr2,
			Is:   false,
		},
		{
			Intf: map[string]string{
				validStr:            validStr,
				validStr + validStr: validStr,
			},
			Is: true,
		},
		{
			Intf: map[string]string{
				validStr:            validStr,
				validStr + validStr: invalidStr,
			},
			Is: false,
		},
	}

	for _, c := range cases {
		if ValidUtf8(c.Intf) != c.Is {
			t.Fatalf("%v expect  %v but %v\n", c.Intf, c.Is, ValidUtf8(c.Intf))
		}
	}
}
