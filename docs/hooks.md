# Pre-Run Hooks

Pre-run hooks allow you to run scripts before your main scripts execute. They're perfect for environment validation, dependency checking, authentication, and setup tasks.

## Quick Start

Create a `.hooks.d` directory in your `TOME_ROOT`:

```bash
mkdir -p $TOME_ROOT/.hooks.d
```

## Hook Types

Hooks come in two flavors:

### Executable Hooks (Separate Process)

These hooks run as independent processes, perfect for validation and checks.

**How to create:**

1. Create a file in `.hooks.d/` **without** `.source` suffix
2. Add shebang: `#!/usr/bin/env bash`
3. Make it executable: `chmod +x .hooks.d/hook-name`
4. Use number prefixes to control order: `00-first`, `10-second`, `20-third`

**Example:** `.hooks.d/00-check-deps`, `.hooks.d/10-validate`

**Best for:**
- Validation checks
- Running external commands
- Dependency verification
- Authentication checks
- Audit logging

### Sourced Hooks (Same Shell Context)

These hooks run in the same shell as your target script, allowing them to modify the environment.

**How to create:**

1. Create a file in `.hooks.d/` **with** `.source` suffix
2. Add number prefix to control order: `05-env.source`, `15-setup.source`
3. No need to make executable (sourcing ignores execute permission)

**Example:** `.hooks.d/05-set-env.source`, `.hooks.d/15-load-functions.source`

**Best for:**
- Setting environment variables
- Defining shell functions
- Modifying PATH
- Loading secrets
- Conditional environment setup

## Environment Variables

All hooks receive these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `TOME_ROOT` | Root directory containing scripts | `/home/user/scripts` |
| `TOME_EXECUTABLE` | Name of the CLI command | `tome-cli` or `kit` |
| `TOME_SCRIPT_PATH` | Full path to script about to run | `/home/user/scripts/deploy` |
| `TOME_SCRIPT_NAME` | Name of script about to run | `deploy` |
| `TOME_SCRIPT_ARGS` | Arguments passed to the script | `"production --force"` |

**Note:** Variables exported by sourced hooks (`.source` suffix) are available to your main script.

## Examples

### Validate AWS Credentials (Executable)

```bash
#!/usr/bin/env bash
# .hooks.d/00-check-aws-creds
# Make executable: chmod +x .hooks.d/00-check-aws-creds

if ! aws sts get-caller-identity &>/dev/null; then
    echo "Error: AWS credentials not configured" >&2
    exit 1
fi

echo "✓ AWS credentials valid"
```

### Set AWS Environment (Sourced)

```bash
# .hooks.d/05-set-aws-env.source
# Note the .source suffix - this will be sourced

# Set AWS environment based on script name
case "$TOME_SCRIPT_NAME" in
    *-prod*)
        export AWS_PROFILE="production"
        export AWS_REGION="us-east-1"
        ;;
    *-staging*)
        export AWS_PROFILE="staging"
        export AWS_REGION="us-west-2"
        ;;
    *)
        export AWS_PROFILE="development"
        export AWS_REGION="us-west-1"
        ;;
esac

echo "✓ AWS environment: $AWS_PROFILE in $AWS_REGION"
```

### Check Dependencies (Executable)

```bash
#!/usr/bin/env bash
# .hooks.d/10-check-deps
# Make executable: chmod +x .hooks.d/10-check-deps

required_tools=("jq" "curl" "aws")

for tool in "${required_tools[@]}"; do
    if ! command -v "$tool" &>/dev/null; then
        echo "Error: $tool not installed" >&2
        exit 1
    fi
done

echo "✓ All dependencies available"
```

### Load Helper Functions (Sourced)

```bash
# .hooks.d/15-functions.source
# Note the .source suffix

# Log with timestamp
log() {
    echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] $*"
}

# Check if in production
is_production() {
    [[ "$AWS_PROFILE" == "production" ]]
}

# Export functions so scripts can use them
export -f log
export -f is_production

echo "✓ Helper functions loaded"
```

### Audit Logging (Executable)

```bash
#!/usr/bin/env bash
# .hooks.d/20-audit-log
# Make executable: chmod +x .hooks.d/20-audit-log

log_file="$TOME_ROOT/.audit-log"
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "$timestamp | $USER | $TOME_SCRIPT_NAME $TOME_SCRIPT_ARGS" >> "$log_file"
echo "✓ Execution logged"
```

### Load Secrets (Sourced)

```bash
# .hooks.d/25-load-secrets.source
# Note the .source suffix

secrets_file="$TOME_ROOT/.secrets"

if [ -f "$secrets_file" ]; then
    source "$secrets_file"
    echo "✓ Secrets loaded"
else
    echo "Warning: $secrets_file not found" >&2
fi
```

## Common Use Cases

### Environment Validation

Validate required environment variables before running scripts:

```bash
#!/usr/bin/env bash
# .hooks.d/00-validate-env

required_vars=("DATABASE_URL" "API_KEY")

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: $var not set" >&2
        exit 1
    fi
done

echo "✓ Environment validated"
```

### Set Default Values

Provide sensible defaults using sourced hooks:

```bash
# .hooks.d/05-defaults.source

export LOG_LEVEL="${LOG_LEVEL:-info}"
export TIMEOUT="${TIMEOUT:-30}"
export RETRY_COUNT="${RETRY_COUNT:-3}"

echo "✓ Defaults set (LOG_LEVEL=$LOG_LEVEL)"
```

### Script-Specific Setup

Use `TOME_SCRIPT_NAME` to customize behavior per script:

```bash
# .hooks.d/10-script-setup.source

# Enable debug mode for test scripts
if [[ "$TOME_SCRIPT_NAME" == test-* ]]; then
    export DEBUG=1
    set -x
fi

# Require confirmation for production scripts
if [[ "$TOME_SCRIPT_NAME" == *-prod ]]; then
    read -p "Running production script. Continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted" >&2
        exit 1
    fi
fi
```

## Error Handling

### Executable Hooks

If an executable hook exits with non-zero status:
- ❌ Main script will NOT execute
- ❌ tome-cli exits with the hook's exit code
- 📋 Error message shows which hook failed

**Example error:**
```
Error: pre-hook failed: 10-check-deps
```

### Sourced Hooks

If a sourced hook fails (bash errors, `set -e` triggered, etc.):
- ❌ Main script will NOT execute
- ❌ tome-cli exits with error
- 📋 Error message shows which hook failed

**Example error:**
```
Error: pre-hook failed: 05-env.source (sourcing failed)
```

## Skipping Hooks

Skip all hooks for a single execution:

```bash
tome-cli --skip-hooks exec my-script
```

This is useful for:
- Testing scripts without hooks
- Performance-critical operations
- Emergency maintenance

## Debugging

See detailed hook execution with the `--debug` flag:

```bash
tome-cli --debug exec my-script
```

This shows:
- Which hooks were discovered
- Hook execution order
- Environment variables set
- Wrapper script path

## Hook Ordering

Hooks execute in **lexicographic order** by filename. Use number prefixes to control execution:

```
.hooks.d/
├── 00-validate-env      # Runs first
├── 05-set-env.source    # Runs second
├── 10-check-deps        # Runs third
├── 15-functions.source  # Runs fourth
└── 20-audit-log         # Runs last
```

**Recommended numbering:**
- `00-09`: Validation and checks
- `10-19`: Setup and configuration
- `20-29`: Logging and monitoring

## Security Considerations

⚠️ **Important Security Notes:**

1. **Executable permissions**: Non-`.source` hooks must be executable to run
2. **Sourced hooks run in same context**: `.source` hooks can modify your shell environment - only use trusted code
3. **Explicit suffix**: The `.source` suffix makes it obvious which hooks modify environment
4. **Hooks run before every script**: Consider the security implications of automatic execution
5. **Hidden directory**: `.hooks.d` is hidden to prevent accidental modification

**Best practices:**
- Review hooks regularly
- Use version control for `.hooks.d/`
- Document what each hook does
- Use `--debug` to verify hook behavior
- Test hooks in non-production first

## Performance

Hooks add minimal overhead:

- ✅ Hook discovery: ~1ms (checks if `.hooks.d/` exists)
- ✅ Wrapper generation: ~1ms (simple string concatenation)
- ⏱️ Hook execution: Depends on your hooks
- 🗑️ Cleanup: Automatic (wrapper script removed after execution)

**Tips for fast hooks:**
- Keep validation hooks simple
- Cache expensive checks when possible
- Use `--skip-hooks` for performance-critical operations
- Sourced hooks are slightly faster than executable hooks (no fork)

## Why Only Pre-Run Hooks?

tome-cli uses `syscall.Exec()` which replaces the process, making post-run hooks technically challenging. Pre-run hooks cover the most common use cases:

✅ Environment validation
✅ Dependency checking
✅ Authentication
✅ Setup and configuration
✅ Audit logging (start)

**Can't do with pre-hooks alone:**
❌ Cleanup after script execution
❌ Capture script exit codes
❌ Post-execution notifications

If you need post-execution behavior, consider:
- Using trap in your scripts: `trap cleanup EXIT`
- Wrapper scripts that call tome-cli then do cleanup
- Process monitoring tools

## Troubleshooting

### Hook not running

**Check:**
1. Is the hook in `.hooks.d/` directory?
2. Is it executable (if not `.source` suffix)? `chmod +x .hooks.d/hook-name`
3. Is the filename correct? (no spaces, proper prefix)
4. Run with `--debug` to see what hooks are discovered

### Sourced hook not modifying environment

**Remember:**
1. File must end with `.source` suffix
2. Use `export` for variables: `export VAR=value`
3. Functions need `export -f`: `export -f function_name`

### Hook fails but script still runs

**This shouldn't happen.** If a hook fails:
- ❌ Script should NOT run
- Check if you're using `--skip-hooks`
- Verify hook exit codes (`exit 1` for errors)

### Want to skip specific hooks

**Options:**
1. Rename hook to something that doesn't match pattern (add `.disabled`)
2. Remove execute permission for executable hooks
3. Move hook out of `.hooks.d/` temporarily
4. Use `--skip-hooks` to skip all hooks

## Migration from Other Systems

If you're coming from:

### Git hooks
- tome hooks are pre-execution, not commit-based
- Use `.hooks.d/` not `.git/hooks/`
- No need for template installation

### Make targets
- Replace prerequisite targets with pre-hooks
- Hooks run automatically, no manual dependencies

### Shell aliases with checks
- Move validation logic into hooks
- Cleaner separation of concerns
- Easier to share across team

## FAQ

**Q: Can I use hooks with aliases?**
A: Yes! Hooks work with both `tome-cli exec` and any aliases you create.

**Q: Do hooks run for tab completion?**
A: No, hooks only run when executing scripts, not during completion.

**Q: Can I have script-specific hooks?**
A: Not yet. Currently only global `.hooks.d/` is supported. Use `TOME_SCRIPT_NAME` to customize behavior per script.

**Q: What shells are supported?**
A: Hooks use `#!/usr/bin/env bash` by default. You can use any interpreter with proper shebang.

**Q: Can hooks modify the script being executed?**
A: No, hooks run before execution but cannot modify the script file itself.

**Q: Are there hook templates?**
A: Not yet, but check the `examples/.hooks.d/` directory for starter templates.

## See Also

- [Writing Scripts](./writing-scripts.md) - Guide to creating tome scripts
- [Completion Guide](./completion-guide.md) - Adding tab completion to scripts
- [Examples](../examples/) - Sample scripts and hooks
