# Completion Guide for tome-cli

This guide explains how to implement custom tab completions for your scripts, providing a seamless command-line experience.

## Table of Contents

- [Overview](#overview)
- [Basic Script Completion](#basic-script-completion)
- [Implementing --complete](#implementing---complete)
- [Completion Format](#completion-format)
- [Examples by Language](#examples-by-language)
- [Advanced Patterns](#advanced-patterns)
- [Debugging Completions](#debugging-completions)

## Overview

tome-cli provides two levels of completion:

1. **Automatic completions** - Completed by tome-cli itself:
   - Built-in commands (`exec`, `help`, `completion`, `alias`)
   - Directory names in your scripts folder
   - Script names in your scripts folder

2. **Script-level completions** - Provided by individual scripts:
   - Script-specific flags and options
   - Dynamic argument values
   - Context-aware suggestions

This guide focuses on implementing script-level completions.

## Basic Script Completion

To enable your script to provide its own completions:

### Step 1: Declare Completion Support

Add `TOME_COMPLETION` anywhere in your script (typically in a comment):

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: my-script [options]

# Rest of your script...
```

### Step 2: Handle the --complete Flag

When tome-cli detects a script with `TOME_COMPLETION`, it will call your script with a `--complete` flag (actually `--completion` based on the code). Handle this flag to output completions:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION

case "${1:-}" in
    --complete)
        # Output your completions here
        echo -e "--help\tShow help message"
        echo -e "--verbose\tEnable verbose mode"
        echo -e "start\tStart the service"
        echo -e "stop\tStop the service"
        exit 0
        ;;
    *)
        # Normal script logic
        echo "Running script..."
        ;;
esac
```

## Implementing --complete

### The --complete Handler Pattern

```bash
case "${1:-}" in
    --complete)
        # Output completions
        # Format: value<TAB>description
        # Use echo -e for tab character (\t)
        echo -e "completion1\tDescription of completion1"
        echo -e "completion2\tDescription of completion2"
        exit 0
        ;;
esac
```

### Important Notes

- Each completion is on its own line
- Format: `value<TAB>description` (tab-separated)
- The description is optional but recommended
- Exit after outputting completions (don't run normal script logic)
- Use `echo -e` to properly interpret the `\t` tab character

## Completion Format

### Basic Format

```
completion-value<TAB>description
```

Examples:

```bash
echo -e "--help\tShow help message"
echo -e "--verbose\tEnable verbose output"
echo -e "production\tProduction environment"
```

### Shell Rendering

Different shells render completions differently:

- **Bash**: Shows completion values (descriptions may not be visible)
- **Zsh**: Shows both values and descriptions
- **Fish**: Shows values with descriptions as hints

### Tips for Good Completions

1. **Keep values short and memorable**: `start`, `stop`, not `start-the-service`
2. **Make descriptions helpful**: Explain what the option does
3. **Order matters**: Put most common options first
4. **Use consistent naming**: Follow CLI conventions (e.g., `--flag-name`)

## Examples by Language

### Bash

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: bash-example <command> [options]

case "${1:-}" in
    --complete)
        # Static completions
        echo -e "start\tStart the service"
        echo -e "stop\tStop the service"
        echo -e "restart\tRestart the service"
        echo -e "status\tCheck service status"
        echo -e "--force\tForce the operation"
        echo -e "--quiet\tSuppress output"
        exit 0
        ;;
    start)
        echo "Starting service..."
        ;;
    stop)
        echo "Stopping service..."
        ;;
    *)
        echo "Usage: $0 <command>" >&2
        exit 1
        ;;
esac
```

### Python

```python
#!/usr/bin/env python3
# TOME_COMPLETION
# USAGE: python-example <command> [options]

import sys

if len(sys.argv) > 1 and sys.argv[1] == '--complete':
    # Output completions
    completions = [
        ("list", "List all items"),
        ("add", "Add a new item"),
        ("remove", "Remove an item"),
        ("--format", "Output format (json, yaml, text)"),
        ("--verbose", "Enable verbose output"),
    ]

    for value, description in completions:
        print(f"{value}\t{description}")

    sys.exit(0)

# Normal script logic
command = sys.argv[1] if len(sys.argv) > 1 else None
if command == "list":
    print("Listing items...")
elif command == "add":
    print("Adding item...")
else:
    print("Unknown command", file=sys.stderr)
    sys.exit(1)
```

### TypeScript/Deno

```typescript
#!/usr/bin/env -S deno run --allow-all
// TOME_COMPLETION
// USAGE: deno-example <command> [options]

if (Deno.args[0] === '--complete') {
  // Output completions
  const completions = [
    ['deploy', 'Deploy the application'],
    ['rollback', 'Rollback to previous version'],
    ['status', 'Check deployment status'],
    ['--environment', 'Target environment'],
    ['--dry-run', 'Preview without executing'],
  ];

  completions.forEach(([value, description]) => {
    console.log(`${value}\t${description}`);
  });

  Deno.exit(0);
}

// Normal script logic
const command = Deno.args[0];
console.log(`Running command: ${command}`);
```

### Ruby

```ruby
#!/usr/bin/env ruby
# TOME_COMPLETION
# USAGE: ruby-example <action> [options]

if ARGV[0] == '--complete'
  # Output completions
  completions = [
    ['backup', 'Create a backup'],
    ['restore', 'Restore from backup'],
    ['list', 'List available backups'],
    ['--destination', 'Backup destination path'],
    ['--compress', 'Compress the backup'],
  ]

  completions.each do |value, description|
    puts "#{value}\t#{description}"
  end

  exit 0
end

# Normal script logic
action = ARGV[0]
puts "Performing action: #{action}"
```

## Advanced Patterns

### Dynamic Completions

Generate completions based on current system state:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: deploy <environment>

case "${1:-}" in
    --complete)
        # List available environments from a config file
        if [ -f "$TOME_ROOT/.config/environments.txt" ]; then
            while IFS= read -r env; do
                echo -e "$env\tDeploy to $env environment"
            done < "$TOME_ROOT/.config/environments.txt"
        else
            # Fallback to static list
            echo -e "development\tDevelopment environment"
            echo -e "staging\tStaging environment"
            echo -e "production\tProduction environment"
        fi
        exit 0
        ;;
    *)
        environment="${1:-}"
        echo "Deploying to $environment..."
        ;;
esac
```

### Context-Aware Completions

Provide different completions based on previous arguments:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: db <command> [options]

case "${1:-}" in
    --complete)
        # If no arguments yet, show commands
        echo -e "backup\tBackup the database"
        echo -e "restore\tRestore the database"
        echo -e "migrate\tRun migrations"
        echo -e "seed\tSeed the database"

        # Common flags
        echo -e "--database\tDatabase name"
        echo -e "--dry-run\tPreview without executing"
        exit 0
        ;;
    backup)
        shift
        case "${1:-}" in
            --complete)
                # Completions specific to backup command
                echo -e "--format\tBackup format (sql, dump)"
                echo -e "--compress\tCompress the backup"
                echo -e "--output\tOutput file path"
                exit 0
                ;;
            *)
                echo "Backing up database..."
                ;;
        esac
        ;;
    *)
        echo "Unknown command" >&2
        exit 1
        ;;
esac
```

### File/Directory Completions

Suggest files or directories:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: process-file <file>

case "${1:-}" in
    --complete)
        # List .json files in current directory
        for file in *.json; do
            [ -f "$file" ] && echo -e "$file\tProcess $file"
        done

        # Or list directories
        for dir in */; do
            [ -d "$dir" ] && echo -e "${dir%/}\tProcess directory: ${dir%/}"
        done

        exit 0
        ;;
    *)
        file="${1:-}"
        echo "Processing file: $file"
        ;;
esac
```

### Reading from External Sources

Fetch completions from APIs or databases:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: cloud <resource> <action>

case "${1:-}" in
    --complete)
        # Fetch available resources from API
        if command -v curl &> /dev/null; then
            curl -s https://api.example.com/resources 2>/dev/null | \
                jq -r '.[] | "\(.name)\t\(.description)"' 2>/dev/null
        else
            # Fallback if curl or jq not available
            echo -e "vm\tVirtual machine"
            echo -e "storage\tStorage bucket"
            echo -e "network\tNetwork configuration"
        fi
        exit 0
        ;;
    *)
        echo "Managing cloud resource..."
        ;;
esac
```

### Caching Completions

Cache expensive completion computations:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: search <query>

CACHE_FILE="$TOME_ROOT/.cache/search-completions"
CACHE_TTL=3600  # 1 hour

case "${1:-}" in
    --complete)
        # Check if cache exists and is fresh
        if [ -f "$CACHE_FILE" ]; then
            cache_age=$(($(date +%s) - $(stat -f %m "$CACHE_FILE" 2>/dev/null || stat -c %Y "$CACHE_FILE")))
            if [ $cache_age -lt $CACHE_TTL ]; then
                cat "$CACHE_FILE"
                exit 0
            fi
        fi

        # Generate completions (expensive operation)
        mkdir -p "$(dirname "$CACHE_FILE")"
        {
            echo -e "recent\tRecent searches"
            echo -e "popular\tPopular searches"
            # ... more completions
        } | tee "$CACHE_FILE"

        exit 0
        ;;
    *)
        query="${1:-}"
        echo "Searching for: $query"
        ;;
esac
```

## Debugging Completions

### Test Completions Manually

Call your script with the --complete flag directly:

```bash
./my-script --complete
```

Expected output:
```
option1	Description of option1
option2	Description of option2
--flag	Description of flag
```

### Check for TOME_COMPLETION Declaration

```bash
grep -l "TOME_COMPLETION" my-script
```

### Verify Tab Character

Ensure you're using actual tabs, not spaces:

```bash
./my-script --complete | cat -A
```

You should see `^I` between values and descriptions (that's a tab).

### Test in Different Shells

```bash
# Test in bash
bash -c 'complete -C "tome-cli completion bash" tome-cli && complete -p tome-cli'

# Test in zsh
zsh -c 'source <(tome-cli completion zsh) && compdef tome-cli'

# Test in fish
fish -c 'tome-cli completion fish | source && complete -C"tome-cli "'
```

### Common Issues

1. **Completions not showing**:
   - Verify `TOME_COMPLETION` is present in the file
   - Check that script handles `--complete` flag
   - Ensure script is executable

2. **Descriptions not visible**:
   - Some shells (bash) don't show descriptions
   - Try in zsh or fish to see descriptions
   - Ensure tab character (`\t`) is used, not spaces

3. **Completions not updating**:
   - Reload shell completions: `source ~/.bashrc`
   - Some shells cache completions aggressively
   - Try in a new shell session

4. **Script execution instead of completions**:
   - Make sure you exit after outputting completions
   - Don't run normal script logic when `--complete` is passed

## Best Practices

1. **Keep it fast**: Completion generation should be quick (<100ms)
2. **Cache when possible**: Cache expensive operations
3. **Fail gracefully**: If completion generation fails, output nothing rather than errors
4. **Test thoroughly**: Test completions in all supported shells
5. **Document your completions**: Mention completion support in your script's help text
6. **Use meaningful descriptions**: Help users understand what each option does
7. **Order by frequency**: Put most common options first
8. **Handle errors silently**: Redirect errors to `/dev/null` in completion mode

## Next Steps

- See [writing-scripts.md](./writing-scripts.md) for general script-writing guidance
- Check out [examples/foo](../examples/foo) for a working completion example
- Read about [environment variables](./writing-scripts.md#using-environment-variables) available in scripts
