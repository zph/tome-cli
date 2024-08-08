package cmd

import (
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

func (s *Script) parse() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	if strings.Contains(string(b), UsageKey) || strings.Contains(string(b), LegacyUsageKey) {
		lines := strings.Split(string(b), "\n")
		var linesStart int
		for idx, line := range lines {
			if strings.Contains(line, UsageKey) || strings.Contains(line, LegacyUsageKey) {
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

		s.usage = strings.TrimSpace(strings.SplitN(lines[linesStart], ":", 2)[1])
		s.help = helpText
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
