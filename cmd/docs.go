/*
Copyright © 2024 Zander Hill	<zander@xargs.io>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Create docs for tome-cli",
	Long: `The docs command generates markdown documentation for the tome-cli command.

	It's hidden from the help menu because it's not a user-facing command.`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := doc.GenMarkdownTree(rootCmd, "docs")
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
