/*
Copyright © 2024 Zander Hill <zander@xargs.io>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: fmt.Sprintf(`To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ %[1]s completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

  PS> %[1]s completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %[1]s completion powershell > %[1]s.ps1
  # and source this file from your PowerShell profile.

Using completions from within child scripts

Once completions discover an executable and non-ignored script,
tome-cli will check if the script has TOME_COMPLETION declared in file.
If so, it will attempt fetch completions from the script itself.

This is accomplished by passing --complete to the script and capturing the output.

The output syntax is newline-separated strings where each line is composed of
a completion (e.g. a flag or argument), a tab character, and a description of the completion.

`, rootCmd.Name()),
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Aliases:               []string{"init"},
	Run: func(cmd *cobra.Command, args []string) {
		stdoutWrappedWriter := &RenameWriter{writer: os.Stdout}
		switch args[0] {
		// In likely case that user has renamed the executable, we need to replace the name in the completion script
		case "bash":
			cmd.Root().GenBashCompletion(stdoutWrappedWriter)
		case "zsh":
			cmd.Root().GenZshCompletion(stdoutWrappedWriter)
		case "fish":
			cmd.Root().GenFishCompletion(stdoutWrappedWriter, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(stdoutWrappedWriter)
		}
	},
}

type RenameWriter struct {
	writer io.Writer
}

func (rw *RenameWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	str = ReplaceNameWithExecutableName(str)
	return rw.writer.Write([]byte(str))
}

func ReplaceNameWithExecutableName(str string) string {
	config := NewConfig()
	exec := stringy.New(config.ExecutableName())
	return strings.ReplaceAll(strings.ReplaceAll(str, "tome-cli", exec.KebabCase().Get()), "tome_cli", exec.SnakeCase().Get())
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
