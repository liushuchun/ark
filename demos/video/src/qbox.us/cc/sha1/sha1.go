package sha1

import (
	sha "crypto/sha1"
)

func Hash(val []byte) []byte {
	h := sha.New()
	h.Write(val)
	return h.Sum(nil)
}
