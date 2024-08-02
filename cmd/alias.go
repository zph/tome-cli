/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"bytes"
	"embed"
	"log"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

type ScriptTemplate struct {
	ExecutableAlias string
	Root            string
}

//go:embed tome-wrapper.sh.tmpl
var content embed.FS

var writePath string

// aliasCmd represents the alias command
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Create an alias wrapper for tome-cli",
	Long: `The alias command allows you to create an alias for the tome command.
This can be useful if you wish to embed common flags like root and executable name and alias the command as a different name.
The generated script uses the executable name and root directory specified in the tome configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := NewConfig()
		s := ScriptTemplate{
			ExecutableAlias: config.ExecutableName(),
			Root:            config.RootDir(),
		}
		t, err := template.ParseFS(content, "tome-wrapper.sh.tmpl")
		// Capture any error
		if err != nil {
			log.Fatalln(err)
		}
		buf := new(bytes.Buffer)
		v, err := cmd.Flags().GetString("write")
		if err != nil {
			panic("Error getting write flag value")
		}
		if v != "" {
			t.Execute(buf, s)
			os.WriteFile(v, buf.Bytes(), 0744)
		} else {
			t.Execute(cmd.OutOrStdout(), s)
		}
	},
}

func init() {
	aliasCmd.Flags().StringVarP(&writePath, "output", "o", "", "Write the alias to a file")
	rootCmd.AddCommand(aliasCmd)
}
