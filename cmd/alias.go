/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	aliasCmd.Flags().StringVarP(&writePath, "write", "w", "", "Write the alias to a file")
	rootCmd.AddCommand(aliasCmd)
}
