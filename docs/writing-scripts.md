# Writing Scripts for tome-cli

This guide covers best practices and patterns for writing scripts that work well with tome-cli.

## Table of Contents

- [Basic Script Structure](#basic-script-structure)
- [Multi-Language Support](#multi-language-support)
- [Help Text Guidelines](#help-text-guidelines)
- [Using Environment Variables](#using-environment-variables)
- [Script Organization](#script-organization)
- [Best Practices](#best-practices)

## Basic Script Structure

Every tome-cli script should follow this basic structure:

```bash
#!/usr/bin/env bash
# USAGE: script-name <required-arg> [optional-arg]
# Brief description of what the script does
#
# More detailed explanation that spans
# multiple lines until the first blank line.
# This entire block becomes the help text.

set -euo pipefail  # Good practice for bash scripts

# Your script logic here
echo "Hello from script-name"
```

### Key Components

1. **Shebang line** (`#!/usr/bin/env bash`): Tells the system how to execute your script
2. **USAGE/SUMMARY comment**: Defines the command syntax and generates help text
3. **Help text**: Continues after USAGE/SUMMARY until first blank line
4. **Script logic**: Your actual implementation

### USAGE vs SUMMARY

tome-cli supports both `USAGE:` and `SUMMARY:` keywords interchangeably:

- **`USAGE:`** - Standard format (recommended)
- **`SUMMARY:`** - Legacy format from original tome (fully supported for backward compatibility)

Both work identically:

```bash
# Option 1: USAGE (recommended for new scripts)
# USAGE: my-script <arg>
# Description here

# Option 2: SUMMARY (legacy, but fully supported)
# SUMMARY: my-script <arg>
# Description here
```

**Best Practice**: Use `USAGE:` for new scripts. Use `SUMMARY:` if migrating from original tome or for consistency with existing scripts that use it.

## Multi-Language Support

tome-cli supports any language with a proper shebang. Here are examples:

### Bash Script

```bash
#!/usr/bin/env bash
# USAGE: bash-example <name>
# Demonstrates a bash script

echo "Hello, $1!"
```

### Python Script

```python
#!/usr/bin/env python3
# USAGE: python-example <number>
# Demonstrates a Python script
#
# This script multiplies the input by 2

import sys

if len(sys.argv) < 2:
    print("Error: number required", file=sys.stderr)
    sys.exit(1)

number = int(sys.argv[1])
print(f"Result: {number * 2}")
```

### TypeScript/Deno Script

```typescript
#!/usr/bin/env -S deno run --allow-all
// USAGE: typescript-example <message>
// Demonstrates a TypeScript/Deno script

const message = Deno.args[0] || "World";
console.log(`Hello, ${message}!`);
```

### Ruby Script

```ruby
#!/usr/bin/env ruby
# USAGE: ruby-example <count>
# Demonstrates a Ruby script

count = ARGV[0]&.to_i || 0
puts "Count: #{count}"
```

### Node.js Script

```javascript
#!/usr/bin/env node
// USAGE: node-example [options]
// Demonstrates a Node.js script

const args = process.argv.slice(2);
console.log('Arguments:', args);
```

## Help Text Guidelines

### Good Help Text

```bash
#!/usr/bin/env bash
# USAGE: deploy <environment> [--force]
# Deploy the application to the specified environment
#
# Arguments:
#   environment  - Target environment (development, staging, production)
#
# Options:
#   --force      - Skip confirmation prompts
#
# Examples:
#   deploy staging
#   deploy production --force
```

### What Makes Good Help Text

1. **Clear usage syntax**: Show required vs optional arguments
2. **Brief description**: One-line summary of what it does
3. **Detailed explanation**: Explain arguments and options
4. **Examples**: Show real usage examples
5. **Stops at blank line**: First blank line ends the help text

### Help Text Parsing Rules

- Must use `USAGE:` or `SUMMARY:` (case-sensitive, with colon)
- Works with any comment style (`#` for most languages, `//` for C-style)
- Should appear in the first ~20 lines of the file
- Continues until the first blank comment line
- The USAGE/SUMMARY line itself is shown as the short help
- Subsequent lines become the detailed help

## Using Environment Variables

tome-cli automatically injects useful environment variables into your scripts:

### Available Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TOME_ROOT` | Absolute path to your scripts directory | `/home/user/my-scripts` |
| `TOME_EXECUTABLE` | Name of the CLI command | `tome-cli` or `kit` |
| `{NAME}_ROOT` | Uppercase executable name + _ROOT | `KIT_ROOT` |
| `{NAME}_EXECUTABLE` | Uppercase executable name + _EXECUTABLE | `KIT_EXECUTABLE` |

### Practical Examples

#### Loading Shared Libraries

```bash
#!/usr/bin/env bash
# USAGE: my-script [options]
# Script that uses shared libraries

# Load common functions from your scripts root
source "$TOME_ROOT/lib/common.sh"

# Now you can use functions from common.sh
common_function_name "$@"
```

#### Referencing Other Scripts

```bash
#!/usr/bin/env bash
# USAGE: orchestrate-tasks
# Runs multiple scripts in sequence

# Call other scripts via the same CLI
"$TOME_ROOT/path/to/setup-script"
"$TOME_ROOT/path/to/main-task"
"$TOME_ROOT/path/to/cleanup-script"
```

#### Using CLI Name in Output

```bash
#!/usr/bin/env bash
# USAGE: help-example
# Shows how to reference the CLI name

echo "Usage: $TOME_EXECUTABLE help-example [options]"
echo ""
echo "Run '$TOME_EXECUTABLE help help-example' for more information"
```

#### Cross-Script Communication

```bash
#!/usr/bin/env bash
# USAGE: store-data <key> <value>
# Store data in a shared cache

CACHE_DIR="$TOME_ROOT/.cache"
mkdir -p "$CACHE_DIR"
echo "$2" > "$CACHE_DIR/$1"
```

## Script Organization

### Recommended Directory Structure

```
my-scripts/
├── .tomeignore          # Files/patterns to ignore
├── lib/                 # Shared libraries
│   └── common.sh
├── db/                  # Database-related scripts
│   ├── backup
│   ├── restore
│   └── migrate
├── deploy/              # Deployment scripts
│   ├── staging
│   └── production
└── utils/               # Utility scripts
    ├── format-code
    └── run-tests
```

### How Organization Affects CLI

Directories become command namespaces:

```bash
# Scripts in root: direct access
my-cli format-code

# Scripts in subdirectories: namespaced
my-cli db backup
my-cli db restore
my-cli deploy production
```

### Organizing Shared Code

Keep shared libraries in a `lib/` directory and exclude them from command listing:

```bash
# .tomeignore
lib/
*.sh   # If you have a lib of .sh files that should be sourced, not executed
```

Then source them in your scripts:

```bash
#!/usr/bin/env bash
source "$TOME_ROOT/lib/common.sh"
```

## Best Practices

### 1. Make Scripts Executable

Always remember to make your scripts executable:

```bash
chmod +x my-script
```

Or for an entire directory:

```bash
chmod +x my-scripts/**/*
```

### 2. Use Proper Error Handling

```bash
#!/usr/bin/env bash
set -euo pipefail  # Exit on error, undefined variables, pipe failures

# Validate required arguments
if [ $# -lt 1 ]; then
    echo "Error: Missing required argument" >&2
    exit 1
fi
```

### 3. Provide Helpful Error Messages

```bash
if [ ! -f "$config_file" ]; then
    echo "Error: Config file not found: $config_file" >&2
    echo "Run '$TOME_EXECUTABLE init' to create a config file" >&2
    exit 1
fi
```

### 4. Support Common Flags

Consider implementing these common flags:

```bash
case "${1:-}" in
    -h|--help)
        # Show help (or let tome-cli handle it)
        $TOME_EXECUTABLE help "$(basename "$0")"
        exit 0
        ;;
    -v|--verbose)
        VERBOSE=true
        shift
        ;;
esac
```

### 5. Use Descriptive Names

Good script names:
- `db-backup` ✓
- `deploy-production` ✓
- `format-yaml` ✓

Poor script names:
- `script1` ✗
- `do-stuff` ✗
- `temp` ✗

### 6. Keep Scripts Focused

Each script should do one thing well. If a script is getting too large:

```bash
# Instead of one giant "deploy" script
deploy-prepare
deploy-build
deploy-upload
deploy-activate

# Or organize them
deploy/prepare
deploy/build
deploy/upload
deploy/activate
```

### 7. Version Control Your Scripts

```bash
# .gitignore for your scripts directory
.cache/
.env
*.log
node_modules/
```

### 8. Document Complex Logic

```bash
#!/usr/bin/env bash
# USAGE: complex-processing <input-file>
# Processes data with multiple transformations

# This script performs the following steps:
# 1. Validates input file format
# 2. Extracts relevant fields
# 3. Applies business logic transformations
# 4. Outputs results in JSON format

process_data() {
    local input="$1"
    # ... implementation
}
```

### 9. Test Your Scripts

```bash
#!/usr/bin/env bash
# USAGE: test-my-feature
# Tests the my-feature script

TOME_ROOT="${TOME_ROOT:-$(dirname "$0")/..}"

# Run your script with test inputs
result=$("$TOME_ROOT/my-feature" "test-input")

# Verify expected output
if [ "$result" = "expected-output" ]; then
    echo "✓ Test passed"
    exit 0
else
    echo "✗ Test failed: got '$result', expected 'expected-output'" >&2
    exit 1
fi
```

### 10. Use .tomeignore Appropriately

```bash
# .tomeignore

# Ignore source files if you have compiled versions
*.ts
*.py

# Ignore test files
*_test
*.test.*

# Ignore hidden files
.*

# Ignore documentation
README.md
*.md

# Ignore specific directories
lib/
test/
.cache/
```

## Common Patterns

### Script with Multiple Subcommands

```bash
#!/usr/bin/env bash
# USAGE: service <command> [options]
# Manage the service
#
# Commands:
#   start    - Start the service
#   stop     - Stop the service
#   restart  - Restart the service
#   status   - Check service status

set -euo pipefail

command="${1:-}"
shift || true

case "$command" in
    start)
        echo "Starting service..."
        # start logic
        ;;
    stop)
        echo "Stopping service..."
        # stop logic
        ;;
    restart)
        "$0" stop
        "$0" start
        ;;
    status)
        echo "Checking service status..."
        # status logic
        ;;
    *)
        echo "Error: Unknown command: $command" >&2
        echo "Run '$TOME_EXECUTABLE help service' for usage" >&2
        exit 1
        ;;
esac
```

### Script with Configuration File

```bash
#!/usr/bin/env bash
# USAGE: configured-script [options]
# Script that reads from a config file

CONFIG_FILE="$TOME_ROOT/.config/my-config.json"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Config file not found: $CONFIG_FILE" >&2
    echo "Run '$TOME_EXECUTABLE init-config' first" >&2
    exit 1
fi

# Read config (example with jq for JSON)
api_key=$(jq -r '.api_key' "$CONFIG_FILE")

# Use the configuration
echo "Using API key: ${api_key:0:8}..."
```

### Script with Progress Output

```bash
#!/usr/bin/env bash
# USAGE: long-running-task
# A task that shows progress

steps=("Initializing" "Processing data" "Validating results" "Cleaning up")

for i in "${!steps[@]}"; do
    current=$((i + 1))
    total=${#steps[@]}
    echo "[$current/$total] ${steps[$i]}..."

    # Do actual work here
    sleep 1
done

echo "✓ Complete!"
```

## Next Steps

- Learn about [adding completions](./completion-guide.md) to your scripts
- See working examples in [examples/](../examples/)
- Read the [migration guide](./migration.md) if coming from tome v1/v2
