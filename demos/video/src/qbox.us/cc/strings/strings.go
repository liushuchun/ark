package strings

func Join(a []string, sep string) []byte {
	if len(a) <= 1 {
		if len(a) == 0 {
			return []byte{}
		}
		return []byte(a[0])
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}
	b := make([]byte, n)
	bp := copy(b, a[0])
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return b
}
