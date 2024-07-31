# tome-cli


A rewrite of `sub` and `tome` but with a different internal implementation and choice of tooling.

# Capabilities

- exec scripts in a directory
- auto-complete script names (https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)
- auto-complete script arguments
- pre-hooks (hooks.d folder will be sourced in order or executed before the real script)
- cross compilation
- test harness for the commands
- tests based on simulated fs
- uncertain if it will allow for sourcing as a core benefit

# Why rewrite?
