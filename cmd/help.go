/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var UsageKey = "USAGE: "

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		TOME_ROOT_DIR := os.Getenv("TOME_ROOT_DIR")
		rootWithArgs := append([]string{TOME_ROOT_DIR}, args...)
		filePath := path.Join(rootWithArgs...)
		b, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		if strings.Contains(string(b), UsageKey) {
			lines := strings.Split(string(b), "\n")
			var linesStart int
			for idx, line := range lines {
				if strings.Contains(line, UsageKey) {
					linesStart = idx
					break
				}
			}

			var helpEnds int
			for idx, line := range lines[linesStart:] {
				if line == "" {
					helpEnds = idx + linesStart
					break
				}
			}
			helpTextLines := lines[linesStart:helpEnds]
			helpText := strings.Join(helpTextLines, "\n")
			fmt.Println(helpText)
		} else {
			fmt.Println("No help available")
		}
		return nil
	},
	ValidArgsFunction: ValidArgsFunctionForScripts,
}

func init() {
	rootCmd.AddCommand(helpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// helpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
