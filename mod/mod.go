package mod

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
)

//此代码出自https://stackoverflow.com/questions/53183356/api-to-get-the-module-name
const (
	RED   = "\033[91m"
	RESET = "\033[0m"
)

func exitf(beforeExitFunc func(), code int, format string, args ...interface{}) {
	beforeExitFunc()
	fmt.Fprintf(os.Stderr, RED+format+RESET, args...)
	//os.Exit(code)
}

func GetModuleName() string {
	goModBytes, err := ioutil.ReadFile("go.mod")
	if err != nil {
		exitf(func() {}, 1, "%+v\n", err)
	}

	modName := modfile.ModulePath(goModBytes)
	//fmt.Fprintf(os.Stdout, "modName=%+v\n", modName)

	return modName
}
