package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/lithammer/dedent"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/viper"
)

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
	return strings.Contains(string(body), "TOME_COMPLETION")
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

// ParseV2 returns the usage and help text for the script
// function aims to return early and perform as little work as possible
// to avoid reading the entire file and stay performant
// with large script folders and files
// Further improvement can be had by detaching the
// parsing of the usage from the parsing of help
// This would benefit the case of rendering help and completions
// where the usage is needed but the help is not
func (s *Script) ParseV2() (string, string, error) {
	log.Debugw("Parsing script", "path", s.path)
	file, err := os.Open(s.path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Parse the script file for the usage and help text
	// Expected structure is:
	// #!/bin/bash
	// # USAGE: script.sh [options] <arg1> <arg2>
	// # This is the help text for the script
	// # It can span multiple lines
	//
	// echo 1

	var usage, help string
	idx := 0

	var helpArr []string

	startsWithComment := regexp.MustCompile(`^[/*\-#]+`)
	for scanner.Scan() {
		t := scanner.Text()
		log.Debugw("Parsing line", "line", t)
		// Skip the shebang line
		if idx == 0 && strings.HasPrefix(t, "#!") {
			log.Debugw("shebang", "line", t)
			idx++
			continue
		}
		// Normally this is the usage line
		if idx == 1 {
			log.Debugw("likely usage", "line", t)
			if !startsWithComment.MatchString(t) {
				usage = ""
				help = ""
				break
			} else {
				withoutCommentChars := strings.TrimLeft(t, "#/-*")
				regexes := []regexp.Regexp{
					*regexp.MustCompile(`(USAGE|SUMMARY):`),
					*regexp.MustCompile(fmt.Sprintf(`(%s|%s)`, regexp.QuoteMeta(`$0`), regexp.QuoteMeta(filepath.Base(s.path)))),
					*regexp.MustCompile(`TOME_[A-Z_]+`), // ignore tome option flags
				}
				for _, r := range regexes {
					withoutCommentChars = r.ReplaceAllLiteralString(withoutCommentChars, "")
				}
				usage = strings.TrimSpace(withoutCommentChars)
				log.Debugw("usage", "usage", usage)
				idx++
			}
		}

		// Scan until we find an empty line
		if startsWithComment.MatchString(t) {
			t2 := strings.TrimSpace(strings.TrimLeft(t, "#/-*"))
			log.Debugw("help line", "line", t2)
			helpArr = append(helpArr, t2)
			idx++
			continue
		} else {
			break
		}
	}
	help = strings.Join(helpArr, "\n")

	return usage, help, nil
}

func (s *Script) parse() error {
	usage, help, err := s.ParseV2()
	if err != nil {
		return err
	}
	s.usage = usage
	s.help = help

	return nil
}

// Usage returns the usage string for the script
// after stripping out the script name or $0
// this is done to reduce visual noise
func (s *Script) Usage() string {
	return dedent.Dedent(s.usage)
}

func (s *Script) Help() string {
	return dedent.Dedent(s.help)
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

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) IgnorePatterns() *gitignore.GitIgnore {
	tomeIgnore := ".tomeignore"
	tomeIgnorePath := filepath.Join(c.RootDir(), tomeIgnore)
	_, err := os.Stat(tomeIgnorePath)
	if err == nil {
		var txt []byte
		txt, err = os.ReadFile(tomeIgnorePath)
		if err != nil {
			fmt.Printf(`Failed to read tome ignore file`)
			os.Exit(1)
		}
		return gitignore.CompileIgnoreLines(strings.Split(string(txt), "\n")...)
	}
	return gitignore.CompileIgnoreLines()
}

func (c *Config) EnvVarWithSuffix(suffix string) (string, bool) {
	prefix := stringy.New(executableName).SnakeCase().Get()
	val := os.Getenv(strings.ToUpper(prefix) + "_" + strings.ToUpper(suffix))
	ok := val != ""

	return val, ok
}

func (c *Config) EnvVarOrViperValue(val string) string {
	v, ok := c.EnvVarWithSuffix(val)
	if ok {
		return v
	}
	return viper.GetViper().GetString(val)
}

func (c *Config) RootDir() string {
	return c.EnvVarOrViperValue("root")
}

func (c *Config) ExecutableName() string {
	return c.EnvVarOrViperValue("executable")
}
