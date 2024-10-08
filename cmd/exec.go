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

		if isExecutableByOwner(fileInfo.Mode()) {
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

	envs := []string{}
	envs = append(envs, fmt.Sprintf("TOME_ROOT=%s", absRootDir))
	envs = append(envs, fmt.Sprintf("TOME_EXECUTABLE=%s", config.ExecutableName()))

	// Inject the named arguments as well
	executableAsEnvPrefix := strings.ToUpper(stringy.New(config.ExecutableName()).SnakeCase().Get())
	envs = append(envs, fmt.Sprintf("%s_ROOT=%s", executableAsEnvPrefix, absRootDir))
	envs = append(envs, fmt.Sprintf("%s_EXECUTABLE=%s", executableAsEnvPrefix, config.ExecutableName()))

	args = append([]string{maybeFile}, maybeArgs...)
	execOrLog(maybeFile, args, envs)
	return nil
}

func execOrLog(arv0 string, argv []string, env []string) {
	if dryRun {
		fmt.Printf("dry run:\nbinary: %s\nargs: %+v\nenv (injected):\n%+v\n", arv0, strings.Join(argv, " "), strings.Join(env, "\n"))
		return
	}
	mergedEnv := append(os.Environ(), env...)

	err := syscall.Exec(arv0, argv, mergedEnv)
	// Exec should create new process, so we should never get here except on error
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "executes a script from tome root",
	Long: dedent.Dedent(`
	Usage: tome-cli exec <path-to> <script> [args...]

	The exec command executes a script file with the provided arguments.

	The exec command will search for the script file in the root directory
	specified in the tome configuration flags or env vars. Paths will be
	joined with the root directory, the intervening directories, and
	the script file name.

	When executed, the script will be become the tome-cli process through
	the syscall.Exec function.

	TOME_ROOT and TOME_EXECUTABLE are injected into the environment as well
	as the executable name as an uppercased snake case string.

	If the executable name is 'kit' the additional environment variables would be:
	KIT_ROOT, KIT_EXECUTABLE.
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
