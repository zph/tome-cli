/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var UsageKey = "USAGE: "

func isExecutableByOwner(mode os.FileMode) bool {
	return mode&0100 != 0
}

type Script struct {
	path  string
	usage string
	help  string
	root  string
}

func (s *Script) parse() error {
	b, err := os.ReadFile(s.path)
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

		s.usage = strings.Split(helpTextLines[0], UsageKey)[1]
		s.help = helpText
	} else {
		fmt.Println("No help available")
	}
	return nil
}

func (s *Script) Usage() string {
	return s.usage
}

func (s *Script) Help() string {
	return s.help
}

func (s *Script) PathSegments() []string {
	return filepath.SplitList(strings.TrimPrefix(strings.TrimPrefix(s.path, s.root), string(filepath.Separator)))
}

func NewScript(path string, root string) *Script {
	s := &Script{path: path, root: root}
	s.parse()
	return s
}

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
		if len(args) == 0 {
			// TODO: print all commands + usage
			allExecutables := []string{}
			fn := func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				if isExecutableByOwner(info.Mode()) {
					allExecutables = append(allExecutables, path)
				}
				return nil
			}
			err := filepath.Walk(TOME_ROOT_DIR, fn)
			if err != nil {
				return err
			}
			for _, executable := range allExecutables {
				s := NewScript(executable, TOME_ROOT_DIR)
				fmt.Printf("%s: %s\n", strings.Join(s.PathSegments(), " "), s.Usage())
			}
		} else {
			rootWithArgs := append([]string{TOME_ROOT_DIR}, args...)
			filePath := path.Join(rootWithArgs...)
			s := NewScript(filePath, TOME_ROOT_DIR)

			fmt.Printf("%s: %s\n", strings.Join(s.PathSegments(), " "), s.Usage())
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
