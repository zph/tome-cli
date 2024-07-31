/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string
var rootDir string
var executableName string

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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "c", "config file (default is local directory or $HOME/.tome-cli.yaml)")
	rootCmd.PersistentFlags().StringVarP(&rootDir, "root", "r", ".", "root directory containing scripts")
	rootCmd.PersistentFlags().StringVarP(&executableName, "executable", "e", "tome-cli", "executable name")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	executablePath, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf(`Unable to determine executable path: %e`, err))
	}
	executableName = filepath.Base(executablePath)
	// init code here

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".tome-cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(executableName)
	}

	viper.SetEnvPrefix(executableName) // will be uppercased automatically
	viper.AutomaticEnv()               // read in environment variables that match
	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	bindFlags(rootCmd, viper.GetViper())

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) RootDir() string {
	return viper.GetString("root")
}
