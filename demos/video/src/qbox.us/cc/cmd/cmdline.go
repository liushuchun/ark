package cmd

import (
	"strings"
)

func IndexNotSpace(line string) int {
	return strings.IndexFunc(line, func(r rune) bool { return r != ' ' && r != '\t' })
}

func ParseCmdline(line string) (args []string) {
	for {
		pos := strings.IndexAny(line, " \t")
		if pos > 0 {
			args = append(args, line[:pos])
		} else if pos < 0 {
			args = append(args, line)
			break
		}
		n := IndexNotSpace(line[pos+1:])
		if n < 0 {
			break
		}
		line = line[pos+n+1:]
	}
	return
}
