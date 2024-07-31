/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zph/tome-cli/cmd"
)

func main() {
	cmd.Execute()
}

var executableName string

// TODO: Change this to the root directory of your project
var rootDir, err = os.LookupEnv("TOME_ROOT_DIR")

func init() {
	executablePath, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf(`Unable to determine executable path: %e`, err))
	}
	executableName = filepath.Base(executablePath)
	// init code here
}
