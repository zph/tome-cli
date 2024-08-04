/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gobeam/stringy"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ExecRunE(cmd *cobra.Command, args []string) error {
	config := NewConfig()
	rootDir := config.RootDir()
	if len(args) == 0 {
		fmt.Println("No file specified")
		os.Exit(1)
	}

	// Try joining one arg segment at a time and return the first one that exists and is executable
	var maybeFile string
	var executable string
	var maybeArgs []string
	for idx, arg := range args {
		// Guard against first cycle through where maybeFile is empty
		// but on subsequent cycles, we want to not double stack the root dir
		var fileRoot string
		if maybeFile == "" {
			fileRoot = rootDir
		} else {
			fileRoot = maybeFile
		}
		maybeFile = path.Join(fileRoot, arg)
		maybeArgs = args[idx+1:]
		fileInfo, err := os.Stat(maybeFile)
		if os.IsNotExist(err) {
			fmt.Printf("File %s does not exist\n", maybeFile)
			continue
		}

		if err != nil {
			fmt.Printf("Error checking file %s: %v\n", maybeFile, err)
			continue
		}

		if fileInfo.IsDir() {
			continue
		}

		if fileInfo.Mode()&0111 != 0 {
			// Found an executable file
			executable = maybeFile
			break
		}
	}
	if executable == "" {
		fmt.Println("No executable file found")
		os.Exit(1)
	}

	absRootDir, err := filepath.Abs(config.RootDir())
	if err != nil {
		fmt.Printf("Error getting absolute path for root dir: %v\n", err)
		os.Exit(1)
	}

	envs := os.Environ()
	envs = append(envs, fmt.Sprintf("TOME_ROOT=%s", absRootDir))
	envs = append(envs, fmt.Sprintf("TOME_EXECUTABLE=%s", config.ExecutableName()))

	// Inject the named arguments as well
	executableAsEnvPrefix := strings.ToUpper(stringy.New(config.ExecutableName()).SnakeCase().Get())
	envs = append(envs, fmt.Sprintf("%s_ROOT=%s", executableAsEnvPrefix, absRootDir))
	envs = append(envs, fmt.Sprintf("%s_EXECUTABLE=%s", executableAsEnvPrefix, config.ExecutableName()))

	args = append([]string{maybeFile}, maybeArgs...)
	err = syscall.Exec(maybeFile, args, envs)
	// Exec should create new process, so we should never get here except on error
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
	return nil
}

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "executes a script from tome root",
	Long: dedent.Dedent(`
		The exec command executes a script file with the provided arguments.

		The exec command will search for the script file in the root directory
	  specified in the tome configuration flags or env vars.

		Scripts are searched for in the root directory and subdirectories and
		then are called with execvp to replace the current process.
		`),
	RunE:              ExecRunE,
	ValidArgsFunction: ValidArgsFunctionForScripts,
}

var dryRun bool

func init() {
	execCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run the exec command")
	viper.BindPFlag("dry-run", execCmd.Flags().Lookup("dry-run"))
	rootCmd.AddCommand(execCmd)
}
