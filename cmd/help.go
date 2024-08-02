/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lithammer/dedent"
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

func (s *Script) HasCompletions() bool {
	body, err := os.ReadFile(s.path)
	if err != nil {
		return false
	}
	return strings.Contains(string(body), "completion")
}

func (s *Script) IsDir() bool {
	fileInfo, err := os.Stat(s.path)
	if err != nil {
		fmt.Printf("Error checking file %s: %v\n", s.path, err)
	}
	return fileInfo.IsDir()
}

func (s *Script) IsExecutable() bool {
	fileInfo, err := os.Stat(s.path)
	if err != nil {
		fmt.Printf("Error checking file %s: %v\n", s.path, err)
	}
	return isExecutableByOwner(fileInfo.Mode())
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

		s.usage = strings.TrimSpace(strings.Split(lines[linesStart], UsageKey)[1])
		s.help = helpText
	} else {
		fmt.Println("No help available")
	}
	return nil
}

// Usage returns the usage string for the script
// after stripping out the script name or $0
// this is done to reduce visual noise
func (s *Script) Usage() string {
	baseUsage := s.usage
	prefixes := []string{"$0", filepath.Base(s.path)}
	for _, prefix := range prefixes {
		baseUsage = strings.TrimPrefix(baseUsage, prefix)
	}
	baseUsage = strings.TrimSpace(baseUsage)
	return dedent.Dedent(baseUsage)
}

func (s *Script) Help() string {
	lines := strings.Split(s.help, "\n")
	var helpTextLines []string
	toTrim := []string{"#", "//", "/\\*", "\\*/", "--"}
	toTrimRegex := regexp.MustCompile(fmt.Sprintf("^(%s)+", strings.Join(toTrim, "|")))
	for _, line := range lines {
		helpTextLines = append(helpTextLines, toTrimRegex.ReplaceAllString(line, ""))
	}
	return dedent.Dedent(strings.Join(helpTextLines, "\n"))
}

func (s *Script) PathWithoutRoot() string {
	return strings.TrimPrefix(strings.TrimPrefix(s.path, s.root), string(filepath.Separator))
}

func (s *Script) PathSegments() []string {
	return strings.Split(s.PathWithoutRoot(), string(filepath.Separator))
}

func (s *Script) PrintUsage() {
	fmt.Printf("%s: %s\n", strings.Join(s.PathSegments(), " "), s.Usage())
}

// PrintHelp prints the full help text for the script
// Help is inclusive of Usage and does not strip out
// the script name or $0
// TODO: consider stripping out leading comment characters such as #, //, etc
func (s *Script) PrintHelp() {
	fmt.Printf("%s\n---\n%s\n", strings.Join(s.PathSegments(), " "), s.Help())
}

func NewScript(path string, root string) *Script {
	s := &Script{path: path, root: root}
	s.parse()
	return s
}

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "help displays the usage and help text for a script",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := NewConfig()
		rootDir := config.RootDir()
		if len(args) == 0 {
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
