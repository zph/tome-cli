/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"io/fs"
	"path"
	"path/filepath"

	"github.com/lithammer/dedent"
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
			allExecutables := []string{}
			fn := func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				if isExecutableByOwner(info.Mode()) && !ignorePatterns.MatchesPath(path) {
					allExecutables = append(allExecutables, path)
				}
				return nil
			}
			// TODO: does not handle symlinks, consider fb symlinkWalk instead
			err := filepath.Walk(rootDir, fn)
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

func init() {
	rootCmd.AddCommand(helpCmd)
}
