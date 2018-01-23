package authutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ParseSuInfo(suInfo string) (uid uint32, appid uint64, err error) {

	pos := strings.Index(suInfo, "/")
	if pos <= 0 {
		err = errors.New("invalid suinfo")
		return
	}
	uidStr, appidStr := suInfo[:pos], suInfo[pos+1:]

	uid64, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		return
	}
	uid = uint32(uid64)
	appid, err = strconv.ParseUint(appidStr, 10, 64)
	if err != nil {
		return
	}

	return
}

func FormatSuInfo(uid uint32, appid uint64) string {

	return fmt.Sprintf("%d/%d", uid, appid)
}
