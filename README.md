# tome-cli

A rewrite of [`sub`](https://github.com/qrush/sub) and [`tome`](https://github.com/toumorokoshi/tome) and my fork of [tome](https://github.com/zph/tome) but with a different internal implementation in order to support:
1. Improved auto-completion
2. Improved usage/help outputs
3. Faster development
4. Testable interfaces

tome-cli transforms a directory of scripts into a unified CLI interface with automatic help generation, tab completion, and support for any programming language.

## Table of Contents

- [Quick Start](#quick-start)
- [Your First Script](#your-first-script)
- [Environment Variables](#environment-variables)
- [Features](#features)
- [Advanced Features](#advanced-features)
- [Troubleshooting](#troubleshooting)
- [Documentation](#documentation)

## Quick Start

### Option 1: Direct Usage (Simple)

Use tome-cli directly by pointing it at your scripts directory:

```bash
# Execute a script
tome-cli --root ~/my-scripts exec myscript arg1 arg2

# Get help for a script
tome-cli --root ~/my-scripts help myscript

# Enable tab completion for your shell
eval "$(tome-cli completion bash)"   # for bash
eval "$(tome-cli completion zsh)"    # for zsh
tome-cli completion fish | source    # for fish
```

### Option 2: Create a Custom CLI (Recommended)

Create a personalized command that embeds your scripts location:

```bash
# Generate a wrapper script called 'kit' that knows about your scripts
tome-cli --root ~/my-scripts --executable kit alias --output ~/bin/kit
chmod +x ~/bin/kit

# Now use your custom command (no need to specify --root anymore)
kit help                    # list all commands
kit myscript arg1 arg2      # run a script
kit completion fish | source  # setup completions for your custom CLI
```

The alias approach is recommended because:
- No need to type `--root` every time
- Shorter command names (e.g., `kit` vs `tome-cli`)
- Completions work with your custom name
- Can be shared with your team

## Your First Script

Create a simple executable script in your scripts directory:

```bash
#!/usr/bin/env bash
# USAGE: hello [name]
# Greets the user by name
#
# Examples:
#   hello Alice
#   hello World

echo "Hello ${1:-World}!"
```

Make it executable: `chmod +x hello`

Now run it: `tome-cli --root ~/my-scripts exec hello Alice`

### Script Requirements

- Must be executable (`chmod +x`)
- Should include a `USAGE:` or `SUMMARY:` line in a comment for automatic help generation
- Can be written in any language (bash, python, ruby, node, deno, etc.)
- Help text is extracted from comments between `USAGE:`/`SUMMARY:` and the first blank line

See [examples/](./examples/) directory for more script examples including completion support.

## Environment Variables

tome-cli automatically injects these environment variables into your scripts:

| Variable | Description | Example |
|----------|-------------|---------|
| `TOME_ROOT` | Absolute path to your scripts directory | `/Users/you/my-scripts` |
| `TOME_EXECUTABLE` | Name of the CLI command being used | `tome-cli` or `kit` |
| `{NAME}_ROOT` | Uppercase version of executable name + _ROOT | `KIT_ROOT` (if executable is `kit`) |
| `{NAME}_EXECUTABLE` | Uppercase version of executable name + _EXECUTABLE | `KIT_EXECUTABLE` |

These are useful for scripts that need to reference other scripts or shared libraries:

```bash
#!/usr/bin/env bash
# Load a shared library from your scripts directory
source "$TOME_ROOT/lib/common.sh"

# Reference the CLI name in help text
echo "Usage: $TOME_EXECUTABLE myscript [options]"
```

See [docs](./docs/tome-cli.md) for expanded instructions

## Features

### Core Features

- **Multi-language support**: Works with any scripting language (bash, python, ruby, node, deno, etc.) via standard shebang (`#!`)
- **Automatic help generation**: Extracts usage text from `USAGE:` comments in your scripts
- **Powerful completions**: Tab-complete commands, subcommands, and even script-specific arguments/flags
- **Custom CLI creation**: Generate personalized CLI commands with embedded configuration via the [alias](./docs/tome-cli_alias.md) feature
- **Script organization**: Organize scripts in directories that become command namespaces
- **Selective execution**: Use `.tomeignore` (gitignore-like syntax) to control which scripts are exposed [example](./cmd/embeds/.tomeignore)

### Auto-Completion Features

tome-cli provides intelligent tab completion for:
- Built-in subcommands (`exec`, `help`, `completion`, etc.)
- Directory names in your scripts folder
- Script names in your scripts folder
- Script-specific flags and arguments (when scripts implement the `--complete` interface with `TOME_COMPLETION`)

See [examples/foo](./examples/foo) for a working example of script-level completion.

## Advanced Features

### .tomeignore File

Control which scripts are exposed as commands using a `.tomeignore` file in your scripts root directory. It uses gitignore-like syntax:

```
# Ignore all TypeScript source files (if you have compiled versions)
*.ts

# Ignore hidden files
.*

# Ignore specific scripts
folder/executable-ignored
```

See [examples/.tomeignore](./examples/.tomeignore) for a working example.

### Script-Level Completions

Your scripts can provide their own tab completions for arguments and flags. This creates a seamless user experience:

1. Add `TOME_COMPLETION` in a comment in your script
2. Handle the `--complete` flag to output completions
3. Format: one completion per line as `value<TAB>description`

Example from [examples/foo](./examples/foo):
```bash
#!/usr/bin/env bash
# TOME_COMPLETION

case $1 in
  --complete)
    echo -e "--help\tShow help message"
    echo -e "--verbose\tEnable verbose output"
    echo -e "start\tStart the service"
    ;;
  *)
    # Normal script logic
    ;;
esac
```

### Pre-Run Hooks

Execute custom scripts before your main scripts run. Perfect for:
- Environment validation and setup
- Dependency checking
- Authentication checks
- Audit logging

Create a `.hooks.d/` directory in your scripts root and add numbered hooks:

```bash
# Executable hook - runs as separate process
.hooks.d/00-check-deps

# Sourced hook - runs in same shell, can modify environment
.hooks.d/05-set-env.source
```

See [docs/hooks.md](./docs/hooks.md) for complete guide with examples.

### Flexible Root Detection

tome-cli determines the scripts root directory from multiple sources (in order of precedence):
1. `--root` CLI flag
2. Environment variable named after the executable (e.g., `PARROT_ROOT` if executable is `parrot`)
3. `TOME_ROOT` environment variable
4. Current directory (default)

This flexibility allows team members to customize locations without changing the CLI tool itself.

## Development Status

### Implemented
- ✅ Execute scripts from any directory structure
- ✅ Auto-complete script names with descriptions
- ✅ Cross-platform compilation via goreleaser
- ✅ E2E test harness with deno
- ✅ Backwards compatibility with original tome
- ✅ Custom CLI aliasing
- ✅ .gitignore-style file filtering
- ✅ Script-level argument completion
- ✅ Environment variable injection
- ✅ Structured logging with levels
- ✅ Generated documentation
- ✅ Pre-run hooks (.hooks.d folder execution)

### Planned
- ⏳ ActiveHelp integration for contextual assistance
- ⏳ Enhanced directory help (show all subcommands in tree)
- ⏳ Improved completion output filtering

## Troubleshooting

### Completions not working

1. **For direct tome-cli usage**: Make sure you've added the completion script to your shell's rc file:
   ```bash
   # Add to ~/.bashrc or ~/.bash_profile
   eval "$(tome-cli completion bash)"

   # Add to ~/.zshrc
   eval "$(tome-cli completion zsh)"

   # Add to ~/.config/fish/config.fish
   tome-cli completion fish | source
   ```

2. **For custom CLI aliases**: Use your alias name in the completion command:
   ```bash
   # If your alias is 'kit'
   eval "$(kit completion bash)"
   ```

3. **After adding completions**: Restart your shell or run `source ~/.bashrc` (or equivalent)

4. **Check script permissions**: Scripts must be executable (`chmod +x script-name`)

### Script not found

1. Verify the script exists and is in the right location:
   ```bash
   ls -la "$TOME_ROOT/path/to/script"
   ```

2. Check if your script is being ignored by `.tomeignore`:
   - Look for a `.tomeignore` file in your `TOME_ROOT`
   - Files matching patterns in `.tomeignore` won't be available as commands
   - By default, hidden files (starting with `.`) are often ignored

3. Ensure the script is executable:
   ```bash
   chmod +x ~/my-scripts/script-name
   ```

4. Use `--debug` flag to see what tome-cli is doing:
   ```bash
   tome-cli --debug --root ~/my-scripts exec script-name
   ```

### Help text not showing / "USAGE:" or "SUMMARY:" not being parsed

The help parser looks for a `USAGE:` or `SUMMARY:` comment near the top of your script:

```bash
#!/usr/bin/env bash
# USAGE: myscript <arg1> [arg2]
# This is the help text
# It continues until the first blank line
#
# This line won't be included (after blank line)
```

Requirements:
- Must use exactly `USAGE:` or `SUMMARY:` (case-sensitive, with colon)
- Should be in a comment within the first ~20 lines
- Help text continues from the line after `USAGE:`/`SUMMARY:` until first blank line
- Works in any language that supports `#` or `//` comments

**Note**: `USAGE:` is recommended for new scripts. `SUMMARY:` is supported for backward compatibility with original tome.

### Tab completion not suggesting arguments

To enable argument/flag completion for your script:

1. Add `TOME_COMPLETION` anywhere in your script (in a comment)
2. Implement a `--complete` flag that outputs completions
3. Format: `completion<TAB>description\n`

Example:
```bash
#!/usr/bin/env bash
# TOME_COMPLETION

case $1 in
  --complete)
    echo -e "--help\tShow help message"
    echo -e "--verbose\tEnable verbose output"
    echo -e "start\tStart the service"
    echo -e "stop\tStop the service"
    ;;
  *)
    # Your script logic here
    ;;
esac
```

See [examples/foo](./examples/foo) for a complete example.

### Permission denied errors

Make sure:
1. Your script is executable: `chmod +x script-name`
2. The script has a valid shebang: `#!/usr/bin/env bash` (or python, ruby, etc.)
3. You have execute permissions on all parent directories

## Documentation

### Getting Started
- [Quick Start](#quick-start) - Get running in minutes
- [Your First Script](#your-first-script) - Create your first script
- [Writing Scripts Guide](./docs/writing-scripts.md) - Comprehensive guide to writing scripts
- [Completion Guide](./docs/completion-guide.md) - Implement custom tab completions
- [Pre-Run Hooks Guide](./docs/hooks.md) - Add validation and setup hooks
- [Migration Guide](./docs/migration.md) - Migrate from original tome/sub

### Core Documentation
- [tome-cli Command Reference](./docs/tome-cli.md) - Complete command-line reference
- [alias Command](./docs/tome-cli_alias.md) - Create custom CLI wrappers
- [completion Command](./docs/tome-cli_completion.md) - Shell completion setup
- [exec Command](./docs/tome-cli_exec.md) - Execute scripts
- [help Command](./docs/tome-cli_help.md) - Display script help

### Examples
- [examples/](./examples/) - Working examples of scripts with various features
- [examples/foo](./examples/foo) - Script with custom completions
- [examples/.tomeignore](./examples/.tomeignore) - Ignore file example

### Additional Resources
- [CHANGELOG.md](./CHANGELOG.md) - Version history and changes
- [Troubleshooting](#troubleshooting) - Common issues and solutions

## Non Features

- Does not support sourcing scripts into shell environment because it adds implementation complexity for other core commands
