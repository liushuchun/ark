package version

import (
	"fmt"
	"io/ioutil"
	"os"
)

var (
	version string      = ""
	pkgName string      = ""
	fPerm   os.FileMode = 0600
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "-version" {
		fmt.Println("version:", version)
		fmt.Println("pkgName:", pkgName)
		os.Exit(0)
	}
	writeFile(".version", version)
	writeFile(".pkg", pkgName)
}

func Version() string {
	return version
}

func PkgName() string {
	return pkgName
}

func writeFile(fname, field string) {
	if field != "" {
		ioutil.WriteFile(fname, []byte(field), fPerm)
	}
}
