package tab

// -----------------------------------------------------------

func Escapes(as []interface{}) []interface{} {

	if dontEscapes(as) {
		return as
	}

	as2 := make([]interface{}, len(as))
	for i, a := range as {
		if s, ok := a.(string); ok {
			as2[i] = Escape(s)
		} else {
			as2[i] = a
		}
	}
	return as2
}

func dontEscapes(as []interface{}) bool {

	for _, a := range as {
		if s, ok := a.(string); ok {
			if needEscape(s) != 0 {
				return false
			}
		}
	}
	return true
}

func Unescapes(as []interface{}) (err error) {

	for _, a := range as {
		if s, ok := a.(*string); ok {
			*s, err = Unescape(*s)
			if err != nil {
				return
			}
		}
	}
	return
}

// -----------------------------------------------------------
