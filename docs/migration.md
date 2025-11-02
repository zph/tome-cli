# Migrating to tome-cli

This guide helps users migrate from the original [tome](https://github.com/toumorokoshi/tome) or [sub](https://github.com/qrush/sub) to tome-cli.

## Table of Contents

- [Why Migrate?](#why-migrate)
- [What's Different](#whats-different)
- [What's the Same](#whats-the-same)
- [Migration Checklist](#migration-checklist)
- [Breaking Changes](#breaking-changes)
- [New Features](#new-features)
- [Troubleshooting Migration](#troubleshooting-migration)

## Why Migrate?

tome-cli is a complete rewrite with several improvements:

1. **Better completions**: Smarter tab completion with description support
2. **Improved help**: Automatic help text extraction from scripts
3. **Faster**: More efficient script discovery and execution
4. **Better tested**: Comprehensive test suite with E2E tests
5. **Active development**: Maintained with new features and bug fixes
6. **Enhanced features**: Script-level completions, better ignore patterns, environment injection

## What's Different

### Installation

**Original tome:**
```bash
# Installed via pip
pip install tome

# Or via shell script
curl -L https://github.com/toumorokoshi/tome/raw/master/install.sh | bash
```

**tome-cli:**
```bash
# Download from releases
# https://github.com/zph/tome-cli/releases

# Or build from source
go build -o tome-cli
```

### Binary Name

**Original tome:**
- Binary name: `tome`

**tome-cli:**
- Binary name: `tome-cli` (but you can create aliases with any name)

### Command Structure

**Original tome:**
```bash
tome script-name args
```

**tome-cli:**
```bash
# Direct execution
tome-cli --root ~/scripts exec script-name args

# Or with alias (recommended)
tome-cli --root ~/scripts --executable tome alias --output ~/bin/tome
tome script-name args
```

### Help Text Format

Both support similar formats, but tome-cli is more flexible:

**Original tome:**
- Used `# summary: description`
- Help text parsing varied

**tome-cli:**
- Supports both `USAGE:` and `SUMMARY:` (for backward compatibility)
- More consistent parsing
- Better multi-line help support

## What's the Same

### Script Organization

Directory structure works the same way:

```
scripts/
├── folder/
│   ├── subscript1
│   └── subscript2
└── top-level-script
```

Commands: `tome folder subscript1`, `tome top-level-script`

### Script Requirements

Scripts still need to be:
- Executable (`chmod +x`)
- Have a shebang (`#!/usr/bin/env bash`, etc.)
- Any language works

### Environment Variables

Both inject similar environment variables:

| Original tome | tome-cli | Description |
|---------------|----------|-------------|
| `TOME_ROOT` | `TOME_ROOT` | Scripts directory |
| N/A | `TOME_EXECUTABLE` | CLI command name |
| `{NAME}_ROOT` | `{NAME}_ROOT` | Uppercase alias + _ROOT |
| N/A | `{NAME}_EXECUTABLE` | Uppercase alias + _EXECUTABLE |

## Migration Checklist

### 1. Install tome-cli

Download from releases or build from source:

```bash
# Download latest release
# https://github.com/zph/tome-cli/releases

# Make it executable
chmod +x tome-cli

# Move to your PATH
mv tome-cli ~/bin/  # or /usr/local/bin/
```

### 2. Create an Alias

Create a `tome` alias to maintain compatibility:

```bash
# Replace original tome with tome-cli
tome-cli --root ~/your-scripts --executable tome alias --output ~/bin/tome
chmod +x ~/bin/tome
```

Now `tome` commands work as before!

### 3. Update Help Text (Optional but Recommended)

Update your scripts to use `USAGE:` format:

**Before (original tome):**
```bash
#!/usr/bin/env bash
# summary: Deploy the application
```

**After (tome-cli - both work):**
```bash
#!/usr/bin/env bash
# USAGE: deploy <environment>
# Deploy the application to the specified environment
#
# This supports multi-line help text until the first blank line.
```

Or keep using `SUMMARY:` - it still works:
```bash
#!/usr/bin/env bash
# SUMMARY: deploy <environment>
# Deploy the application
```

### 4. Update Ignore Patterns

**Original tome:**
- Used `.tomerc` file

**tome-cli:**
- Uses `.tomeignore` file (gitignore syntax)

Migrate your ignore patterns:

```bash
# .tomerc (original tome)
[ignore]
patterns = *.pyc, __pycache__, test_*

# .tomeignore (tome-cli)
*.pyc
__pycache__/
test_*
```

### 5. Update Shell Completions

**Original tome:**
```bash
# In ~/.bashrc
source ~/tome/completion/bash/tome.bash
```

**tome-cli:**
```bash
# In ~/.bashrc
eval "$(tome completion bash)"

# Or with your custom alias
eval "$(~/bin/tome completion bash)"
```

### 6. Test Your Scripts

Run your existing scripts with tome-cli:

```bash
# List all commands (verify everything is detected)
tome help

# Test a specific script
tome your-script-name

# Check completions work
tome <TAB><TAB>
```

### 7. Update CI/CD and Documentation

Update any references to tome installation and usage in:
- CI/CD pipelines
- Documentation
- Setup scripts
- Team onboarding guides

## Breaking Changes

### 1. Different Binary Name

**Impact**: Scripts or tools that call `tome` directly will need updating.

**Solution**:
```bash
# Create an alias with the original name
tome-cli --root ~/scripts --executable tome alias --output ~/bin/tome
```

### 2. Different Ignore File

**Impact**: `.tomerc` is not used; `.tomeignore` is used instead.

**Solution**: Convert your `.tomerc` patterns to `.tomeignore` (gitignore syntax).

### 3. Exec Command Required (Without Alias)

**Impact**: Direct usage requires `exec` subcommand.

**Solution**:
```bash
# Instead of: tome script-name
tome-cli --root ~/scripts exec script-name

# Or create an alias to avoid this
tome-cli --root ~/scripts --executable tome alias --output ~/bin/tome
# Now: tome script-name works
```

### 4. No Built-in Config File

**Impact**: Original tome used `.tomerc` for configuration.

**Solution**: tome-cli uses CLI flags or environment variables:
```bash
# Use environment variables
export TOME_ROOT=~/scripts
tome-cli exec script-name

# Or create an alias with embedded config
tome-cli --root ~/scripts --executable tome alias --output ~/bin/tome
```

## New Features

tome-cli adds features not available in original tome:

### 1. Script-Level Completions

Your scripts can provide their own tab completions:

```bash
#!/usr/bin/env bash
# TOME_COMPLETION
# USAGE: my-script <command>

case "${1:-}" in
    --complete)
        echo -e "start\tStart the service"
        echo -e "stop\tStop the service"
        exit 0
        ;;
    *)
        # Normal script logic
        ;;
esac
```

See [completion-guide.md](./completion-guide.md) for details.

### 2. Better Environment Injection

tome-cli injects more environment variables:

```bash
#!/usr/bin/env bash
# Access the CLI executable name
echo "Run: $TOME_EXECUTABLE help for more info"

# Access the root directory
source "$TOME_ROOT/lib/common.sh"
```

### 3. Enhanced Help Generation

Multi-line help text with better formatting:

```bash
#!/usr/bin/env bash
# USAGE: my-script <arg1> [arg2]
# This is a detailed description
# that spans multiple lines.
#
# Examples:
#   my-script value1
#   my-script value1 value2
```

### 4. Flexible Alias System

Create multiple custom CLIs from the same scripts:

```bash
# Create different aliases for different teams
tome-cli --root ~/scripts --executable team-a alias --output ~/bin/team-a
tome-cli --root ~/scripts --executable team-b alias --output ~/bin/team-b
```

### 5. Better Ignore Patterns

Gitignore-like syntax with more flexibility:

```
# .tomeignore
*.bak
*.tmp
test_*/
lib/*.sh  # Ignore library files
```

## Troubleshooting Migration

### Scripts Not Found After Migration

**Symptom**: `tome script-name` says script not found.

**Possible causes**:
1. Script not executable: `chmod +x script-name`
2. Script being ignored: Check `.tomeignore`
3. Wrong root directory: Verify `$TOME_ROOT` or alias configuration

**Solution**:
```bash
# Check what scripts are detected
tome help

# Use debug mode to see what's happening
tome-cli --debug --root ~/scripts exec script-name
```

### Completions Not Working

**Symptom**: Tab completion doesn't show your scripts.

**Possible causes**:
1. Completions not installed
2. Using wrong completion command
3. Shell not reloaded

**Solution**:
```bash
# Reinstall completions with your alias name
eval "$(tome completion bash)"

# Reload shell
source ~/.bashrc

# Or start a new shell session
exec bash
```

### Help Text Not Showing

**Symptom**: `tome help script-name` shows nothing.

**Possible causes**:
1. No `USAGE:` or `SUMMARY:` in script
2. Help text format incorrect
3. Blank line after shebang missing

**Solution**:
```bash
#!/usr/bin/env bash
# USAGE: script-name
# Description here
#
# (blank line ends help text)

# Rest of script
```

### Different Behavior from Original tome

**Symptom**: Scripts behave differently under tome-cli.

**Possible causes**:
1. Different environment variables
2. Different execution context
3. Different PATH or environment

**Debug approach**:
```bash
# Add debugging to your script
#!/usr/bin/env bash
# USAGE: debug-script

echo "TOME_ROOT: $TOME_ROOT"
echo "TOME_EXECUTABLE: $TOME_EXECUTABLE"
echo "PATH: $PATH"
env | grep TOME
```

### Performance Issues

**Symptom**: tome-cli slower than original tome.

**Unlikely but possible solutions**:
1. Check for large `.tomeignore` patterns
2. Verify scripts directory isn't huge
3. Use `--debug` to see where time is spent

```bash
time tome-cli --debug --root ~/scripts exec script-name
```

## Getting Help

If you encounter issues during migration:

1. **Check the documentation**:
   - [README.md](../README.md) - General usage
   - [writing-scripts.md](./writing-scripts.md) - Script writing guide
   - [completion-guide.md](./completion-guide.md) - Completion implementation

2. **Enable debug mode**:
   ```bash
   tome-cli --debug --root ~/scripts exec script-name
   ```

3. **Compare with examples**:
   - Look at [examples/](../examples/) directory
   - Test with the example scripts first

4. **Report issues**:
   - Open an issue on GitHub with details:
     - Original tome version
     - tome-cli version
     - Script that's not working
     - Error messages and debug output

## Migration Success Stories

After migration, you'll have:

- ✅ Better tab completions with descriptions
- ✅ Consistent help text formatting
- ✅ Script-level completions
- ✅ More reliable execution
- ✅ Better debugging tools
- ✅ Active maintenance and updates
- ✅ Backward compatibility with existing scripts

Most migrations are straightforward - create an alias and update completions. Your scripts should work as-is!
