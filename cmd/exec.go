/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func ValidArgsFunctionForScripts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	config := NewConfig()
	rootDir := config.RootDir()

	if debug {
		cobra.CompDebugln(fmt.Sprintf(`completion: args=%+v, toComplete=%s`, args, toComplete), true)
	}
	// If we have an executable file in the path, we're working on completions for that script itself via --completion
	var argsAccumulator []string
	// Iteration must exit on first matching executable file or it breaks invariants of code
	for _, arg := range args {
		// __complete is passed as an internal directive
		if arg == "__complete" {
			continue
		}
		argsAccumulator = append(argsAccumulator, arg)
		joint := path.Join(append([]string{rootDir}, argsAccumulator...)...)
		f, err := os.Stat(joint)
		if debug {
			cobra.CompDebugln(fmt.Sprintf(`completion: joint=%s`, joint), true)
			cobra.CompDebugln(fmt.Sprintf(`completion: executable=%s`, f.Mode()), true)
		}
		if err == nil && isExecutableByOwner(f.Mode()) {
			s := NewScript(joint, rootDir)
			// We have an executable file in the path
			// Handle completion for the script itself via --completion

			/*
				Check if the script contains word --completion
			*/
			if debug {
				cobra.CompDebugln(fmt.Sprintf(`completion: hasCompletions=%t`, s.HasCompletions()), true)
			}
			if !s.HasCompletions() {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			/*
				Extract the completion values from the script
			*/
			// Execute the joint path as a shell script
			completionFlag := []string{"--completion"}
			output, err := exec.Command(joint, completionFlag...).Output()
			if debug {
				cobra.CompDebugln(fmt.Sprintf(`completion: output=%s`, output), true)
			}
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			// Split the output into lines
			lines := strings.Split(string(output), "\n")

			if debug {
				cobra.CompDebugln(fmt.Sprintf(`completion: lines=%s`, lines), true)
			}
			// Remove empty lines
			var completions []string
			for _, line := range lines {
				if debug {
					cobra.CompDebugln(fmt.Sprintf(`completion: line=%+v`, line), true)
				}
				if line != "" {
					completions = append(completions, line)
				}
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// Otherwise we're completing the path to the script
	fullPathSegments := append([]string{rootDir}, args...)
	fullPath := path.Join(fullPathSegments...)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	if len(entries) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var toCompleteEntries []string
	if toComplete == "" {
		for _, entry := range entries {
			toCompleteEntries = append(toCompleteEntries, entry.Name())
		}
	} else {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), toComplete) {
				toCompleteEntries = append(toCompleteEntries, entry.Name())
			}
		}
	}

	var executableOrDirectories []string
	for _, entry := range toCompleteEntries {
		if strings.HasSuffix(entry, "/") {
			executableOrDirectories = append(executableOrDirectories, entry)
		} else {
			fullPathWithEntry := path.Join(fullPath, entry)
			s := NewScript(fullPathWithEntry, rootDir)
			if s.IsDir() {
				executableOrDirectories = append(executableOrDirectories, entry+"\tdirectory")
			} else if s.IsExecutable() {
				if debug {
					cobra.CompDebugln(fmt.Sprintf(`completion: fullPath=%s, entry=%s, %+v`, fullPath, entry, s), true)
				}
				executableOrDirectories = append(executableOrDirectories, entry+"\t"+s.Usage())
			}
		}
	}
	return executableOrDirectories, cobra.ShellCompDirectiveNoFileComp
}

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
	Use:               "exec",
	Short:             "executes a script",
	RunE:              ExecRunE,
	ValidArgsFunction: ValidArgsFunctionForScripts,
}

func init() {
	rootCmd.AddCommand(execCmd)
}
