/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gobeam/stringy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootDir string
var executableName string
var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tome-cli",
	Short: "A cli tool to manage scripts as a set of subcommands",
	Long: `tome-cli is a command-line tool that allows you to manage scripts as a set of subcommands.

It succeeds sub and tome as a third generation that borrows much of it's design from those projects.

It provides a convenient way to organize and execute scripts within a project.
By loading the context of the full git repository, tome-cli enables you to access and execute scripts specific to your project. It leverages the power of Cobra, a CLI library for Go, to provide a user-friendly and efficient command-line interface.
For more information and usage examples, please refer to the documentation and examples provided in the repository.`,
	// Bare command is `exec` and it requires at least one argument
	Args: cobra.MinimumNArgs(1),
	RunE: ExecRunE,
	// cobra automatically injects subcommands into the custom shell completion :)
	// So in this case we have subcommands mixed with scripts auto-completion
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
	log := createLogger("initConfig", rootCmd.OutOrStderr())
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
