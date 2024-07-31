/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func ValidArgsFunctionForScripts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	rootDir, ok := os.LookupEnv("TOME_ROOT_DIR")
	if !ok {
		fmt.Println("TOME_ROOT_DIR not set")
		os.Exit(1)
	}

	fullPathSegments := append([]string{rootDir}, args...)
	fullPath := path.Join(fullPathSegments...)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	if len(entries) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// TODO: determine how to auto-complete flags for dynamic subcommands
	// TODO: how do we do completions for arbitrary binary name? right now it uses tome-cli
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
			fileInfo, err := os.Stat(path.Join(fullPath, entry))
			if err != nil {
				fmt.Printf("Error checking file %s: %v\n", entry, err)
				continue
			}
			if fileInfo.IsDir() {
				executableOrDirectories = append(executableOrDirectories, entry+"\tdirectory")
			} else if fileInfo.Mode()&0111 != 0 {
				// Found an executable file
				b, err := os.ReadFile(path.Join(fullPath, entry))
				if err != nil {
					return nil, cobra.ShellCompDirectiveError
				}
				lines := strings.Split(string(b), "\n")
				var usage string
				for _, line := range lines {
					if strings.Contains(line, UsageKey) {
						usage = strings.Split(line, UsageKey)[1]
						break
					}
				}
				executableOrDirectories = append(executableOrDirectories, entry+"\t"+usage)
			}
		}
	}
	return executableOrDirectories, cobra.ShellCompDirectiveNoFileComp
}

func ExecRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Println("No file specified")
		os.Exit(1)
	}

	// Try joining one arg segment at a time and return the first one that exists and is executable
	var maybeFile string
	var executable string
	var maybeArgs []string
	for idx, arg := range args {

		rootDir, ok := os.LookupEnv("TOME_ROOT_DIR")
		if !ok {
			fmt.Println("TOME_ROOT_DIR not set")
			os.Exit(1)
		}
		maybeFile = path.Join(rootDir, maybeFile, arg)
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

	args = append([]string{maybeFile}, maybeArgs...)
	err := syscall.Exec(maybeFile, args, os.Environ())
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE:              ExecRunE,
	ValidArgsFunction: ValidArgsFunctionForScripts,
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// execCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// execCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
