/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootDir string
var executableName string
var debug bool

var (
	replaceHyphenWithCamelCase = false
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tome-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Bare command is `exec` and it requires at least one argument
	Args: cobra.MinimumNArgs(1),
	RunE: ExecRunE,
	// TODO: validate that auto-completion includes scripts currently only includes subcommands
	ValidArgsFunction: ValidArgsFunctionForScripts,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Disable the builtin help subcommand
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// We cannot directly bind viper to the rootCmd because
	// the flag default values will override anything in config file :-/
	// Instead we tried bindFlags from https://github.com/carolynvs/stingoftheviper/blob/main/main.go#L111-L128
	// But that seems to break the environment variable binding
	rootCmd.PersistentFlags().StringVarP(&rootDir, "root", "r", ".", "root directory containing scripts")
	rootCmd.PersistentFlags().StringVarP(&executableName, "executable", "e", "", "executable name")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug logs")
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("executable", rootCmd.PersistentFlags().Lookup("executable"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("author", "Zander Hill <zander@xargs.io>")
	viper.SetDefault("license", "mit")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	log := createLogger("initConfig", rootCmd.OutOrStdout())
	v := viper.GetViper()
	var err error
	rootDir, err = filepath.Abs(rootDir)
	log.Debug("rootDir", rootDir)
	if err != nil {
		panic(fmt.Sprintf(`Unable to determine absolute path for root directory: %e`, err))
	}

	log.Debugw("executableName from flags", "var", executableName)
	if executableName == "" {
		executablePath, err := os.Executable()
		if err != nil {
			panic(fmt.Sprintf(`Unable to determine executable path: %e`, err))
		}
		executableName = filepath.Base(executablePath)
		log.Debug("executableName from binary", executableName)
	}

	ex := stringy.New(executableName).SnakeCase().Get()
	log.Debugw("env prefix", "ex", ex)
	// TOME will be used consistently as the prefix for environment variables
	// we will overlay env prefix with the executable name and use it if
	// it is set to support multiple instances of the cli
	v.SetEnvPrefix("TOME") // will be uppercased automatically
	v.AutomaticEnv()       // read in environment variables that match
}

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
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
