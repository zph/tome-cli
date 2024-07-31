# tome-cli


A rewrite of [`sub`](https://github.com/qrush/sub) and [`tome`](https://github.com/toumorokoshi/tome) and my fork of [tome](https://github.com/zph/tome) but with a different internal implementation and choice of tooling.

# Interface

```
tome exec path to file
tome help path to <TAB>

# Print out completions for zsh | fish | bash
tome completion zsh

All of which
```
# Capabilities

- exec scripts in a directory
- auto-complete script names (https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)
- auto-complete script arguments
- pre-hooks (hooks.d folder will be sourced in order or executed before the real script)
- cross compilation
- test harness for the commands
- tests based on simulated fs
- uncertain if it will allow for sourcing as a core benefit, if so it will be under dedicated
  command (source)
- injects TOME_ROOT into tooling as env var (but could inject as MY_COMMAND_ROOT if users have
  multiple tome-cli running)
- Determines root folder based on:
    1. cli flag
    2. env var named after binary (eg auto detect and use PARROT_ROOT if alias name is parrot)
- Respects a .gitignore type file in root of project
- Add description to command completions
- Use https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md#creating-your-own-completion-command
  to override and rename from tome-cli to whatever the binary name is on their system

# Why rewrite?

Because I'm proficient in golang and see architectural choices in the other implementations that
don't support my design of this tool.

# Alternate names

kit
edc or eds (everyday scripting)
grimoire
