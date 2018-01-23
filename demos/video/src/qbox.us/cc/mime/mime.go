package mime

import (
	"encoding/json"
	"strings"

	. "github.com/qiniu/ctype"
)

const (
	validMimeCtype = ALPHA | DIGIT | BLANK | UNDERLINE | SUB | ADD | DOT | DIV | EQ | COMMA | SEMICOLON
)

// custom mimeType
func init() {
	generateRmimes()
}

func generateRmimes() {

	rmineCfg := make(map[string][]string)
	err := json.Unmarshal([]byte(rmime_conf), &rmineCfg)
	if err != nil {
		panic(err)
	}
	for k, v := range rmineCfg {
		rmimes[k] = v[0]
		for _, ext := range v {
			if mtype, ok := mimes[ext]; !ok {
				mimes[ext] = k
			} else {
				c1 := strings.HasPrefix(mtype, "applicaiton/")
				c2 := strings.HasPrefix(k, "applicaiton/")
				c3 := mtype > k
				if (c1 && c2 && c3) || (!c1 && !c2 && c3) || (!c1 && c2) {
					mimes[ext] = k
				}
			}
		}
	}

	rmimes["application/octet-stream"] = ""
}

func ContentType(ext string) string {

	return mimes[strings.ToLower(ext)]
}

func Suffix(mimeType string) string {

	return rmimes[strings.ToLower(mimeType)]
}

/*
	mimetype合法性检查
	1. 目前所有的mimetype包括如下字符集：a-z, A-Z, 0-9, ., +, /, -
	2. 因为mimetype日后还会增加，为保证兼容将来的增加如下字符“;=,_”和空格
	3. 目前mimetype最大长度为79, 为保证兼容将来的这里限制最大长度为200
	4. mimetype为空这里认为合法，因为如果为空后面会进一步检测并设置
	5. mimetype可以包括空格，例如：“text/plain; charset=iso-8859-1”
	6. 禁止转义，如"%0A"表示"\n"，所以禁止出现"%"
*/
func IsValidMimeType(mimeType string) bool {
	if mimeType == "" {
		return true
	}
	return len(mimeType) <= 200 && IsType(validMimeCtype, mimeType)
}

// =======================================================================

// 用户自定义的meta key 支持 字母、数字、下划线、减号，长度小于等于50
func IsValidMetaKey(key string) bool {
	return len(key) <= 50+len("x-qn-meta-") && IsType(XMLSYMBOL_NEXT_CHAR, key)
}
