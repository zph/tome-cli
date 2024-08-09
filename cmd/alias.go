/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"bytes"
	"embed"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

type ScriptTemplate struct {
	ExecutableAlias string
	Root            string
}

//go:embed embeds/tome-wrapper.sh.tmpl
//go:embed embeds/.tomeignore
var content embed.FS

var writePath string

// aliasCmd represents the alias command
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Create an alias wrapper for tome-cli",
	Long: `The alias command allows you to create an alias for the tome command.

The alias command allows you to create an alias for the tome command.

An alias is a shell script that embeds common flags like root and executable
name and can be stored as an alternate name.

The generated script uses the executable name and root directory specified in the tome configuration file.

To use it:

	The following command will create an alias script in the ~/bin directory named 'kit'
	which embeds the root directory and executable name so that 'kit' can be used in normal
	circumstances with no flags or environment variables.

  $> tome-cli --root $PWD/examples --executable kit alias --output ~/bin/kit

Read the template script 'tome-wrapper.sh.tmpl' for more information on how the alias is created
`,
	Run: func(cmd *cobra.Command, args []string) {
		config := NewConfig()
		s := ScriptTemplate{
			ExecutableAlias: config.ExecutableName(),
			Root:            config.RootDir(),
		}
		t, err := template.ParseFS(content, "embeds/tome-wrapper.sh.tmpl")
		// Capture any error
		if err != nil {
			log.Fatal(err)
		}
		buf := new(bytes.Buffer)
		v, err := cmd.Flags().GetString("output")
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
