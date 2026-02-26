/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/lithammer/dedent"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
)

var UsageKey = "USAGE: "
var LegacyUsageKey = "SUMMARY: "

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "help displays the usage and help text for a script",
	Long: dedent.Dedent(`
	The help command extracts the usage and help text from a script file and displays it to the user.

	Help text is extracted from the script file by searching for the first line that includes "USAGE: ".

	When printing long form help text, the help command will print the help text from the script file
  starting from the line after the "USAGE: " line and ending on the first blank line.

	Example: (more can be found in the examples directory)

	#!/bin/bash
	# USAGE: script.sh [options] <arg1> <arg2>
	# This is the help text for the script
	# It can span multiple lines
	# and will be displayed to the user when they run "tome-cli help script.sh"

	echo 1

	In this example the USAGE line is "USAGE: script.sh [options] <arg1> <arg2>"
	The help text is the lines following the USAGE line until the first blank line.
	`),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := NewConfig()
		rootDir := config.RootDir()
		ignorePatterns := config.IgnorePatterns()
		if len(args) == 0 {
			allExecutables, err := collectExecutables(rootDir, ignorePatterns)
			if err != nil {
				return err
			}
			for _, executable := range allExecutables {
				s := NewScript(executable, rootDir)
				s.PrintUsage()
			}
		} else {
			rootWithArgs := append([]string{rootDir}, args...)
			filePath := path.Join(rootWithArgs...)
			s := NewScript(filePath, rootDir)
			s.PrintHelp()
		}
		return nil
	},
	ValidArgsFunction: ValidArgsFunctionForScripts,
}

// collectExecutables walks rootDir and returns paths to all executable files,
// resolving symlinks to check the target's properties.
// SYMLINK-001, SYMLINK-002: symlinked executables are included.
// SYMLINK-003: broken symlinks are skipped without error.
func collectExecutables(rootDir string, ignorePatterns *gitignore.GitIgnore) ([]string, error) {
	var allExecutables []string
	fn := func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// SYMLINK-001, SYMLINK-002: resolve symlinks to check target properties
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, statErr := os.Stat(p)
			if statErr != nil {
				// SYMLINK-003: skip broken symlinks
				return nil
			}
			info = resolved
		}
		if info.IsDir() {
			return nil
		}
		if isExecutableByOwner(info.Mode()) && !ignorePatterns.MatchesPath(p) {
			allExecutables = append(allExecutables, p)
		}
		return nil
	}
	err := filepath.Walk(rootDir, fn)
	return allExecutables, err
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
