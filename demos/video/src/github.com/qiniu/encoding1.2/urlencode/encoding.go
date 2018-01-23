package urlencode

import (
	"net/url"
)

func UrlEncoding(in string) string {

	return url.QueryEscape(Escape(in))
}

func UrlUnEncoding(in string) (string, error) {

	urlStr, err := url.QueryUnescape(in)
	if err != nil {
		return "", err
	}
	return UnEscape(urlStr), err
}

/*
 *   转义字符为 '@'
 *   句首的  '/' '@'  需要转义
 *   '/'后的 '/' '@'  需要转义
 */
func Escape(in string) string {

	nEsc := 0
	needEsc := true
	for i := 0; i < len(in); i++ {
		if needEsc && (in[i] == '/' || in[i] == '@') {
			nEsc++
		}
		needEsc = (in[i] == '/')
	}
	if nEsc == 0 {
		return in
	}

	var retBytes = make([]byte, 0, len(in)+nEsc)
	needEsc = true
	for i := 0; i < len(in); i++ {
		if needEsc && (in[i] == '/' || in[i] == '@') {
			retBytes = append(retBytes, '@')
		}
		retBytes = append(retBytes, in[i])
		needEsc = (in[i] == '/')
	}
	return string(retBytes)
}

func UnEscape(in string) string {

	if len(in) < 2 {
		return in
	}
	nUnesc := 0
	var preCh = in[0]
	needUnEsc := preCh == '@'
	for i := 1; i < len(in); i++ {
		if needUnEsc && (in[i] == '/' || in[i] == '@') {
			nUnesc++
		}
		needUnEsc = (preCh == '/' && in[i] == '@')
		preCh = in[i]
	}
	if nUnesc == 0 {
		return in
	}

	var retBytes = make([]byte, 0, len(in)-nUnesc)

	preCh = in[0]
	needUnEsc = preCh == '@'
	for i := 1; i < len(in); i++ {
		if !(needUnEsc && (in[i] == '/' || in[i] == '@')) {
			retBytes = append(retBytes, preCh)
		}
		needUnEsc = (preCh == '/' && in[i] == '@')
		preCh = in[i]
	}
	retBytes = append(retBytes, preCh)

	return string(retBytes)
}
