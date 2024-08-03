# tome-cli


A rewrite of [`sub`](https://github.com/qrush/sub) and [`tome`](https://github.com/toumorokoshi/tome) and my fork of [tome](https://github.com/zph/tome) but with a different internal implementation in order to support:
1. Improved auto-completion
2. Improved usage/help outputs
3. Faster development
4. Testable interfaces

# Interface

```
export TOME_ROOT=examples
tome-cli exec path to file
tome-cli help path to <TAB>
tome-cli completion fish | source
tome-cli alias --write kit

# shorthand syntax via bash wrapper script
tome-cli --executable kit alias --output ~/bin/kit

# further uses of kit script have embedded values for TOME_ROOT and TOME_EXECUTABLE
kit completion fish | source
kit path to file
kit pat<TAB>


# Print out completions for zsh | fish | bash
tome completion zsh
```

# Capabilities

- [x] exec scripts in a directory
- [x] auto-complete script names (https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)
- [x] include description into auto-complete if shell supports it
- [x] cross compilation via goreleaser
- [o] test harness for the commands
  - [x] e2e tests wrapped in deno unit tests
  - [ ] Add tests for backwards compatibility with prior tome
  - [ ] cobra testing for commands
- [x] supports aliasing tool to shorthand name
- [x] Determines root folder based on:
        1. cli flag
        2. env var named after binary (eg auto detect and use PARROT_ROOT if alias name is parrot)
- [x] Use https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md#creating-your-own-completion-command
      to override and rename from tome-cli to whatever the binary name is on their system
- [x] add tome compatibility layer for transition or include shell script shim
- [ ] Add instructions to README
- [ ] Generate a docs folder for more full instructions (https://umarcor.github.io/cobra/#generating-markdown-docs-for-your-own-cobracommand)
- [ ] Setup changelog tooling ([chglog init](https://github.com/goreleaser/chglog))
- [ ] Respects a .gitignore type file in root of project to determine what to complete/execute (https://github.com/sabhiram/go-gitignore)
- [x] auto-complete script arguments (scripts that include the text TOME_COMPLETION which are tab completed will try to get autocompletes from the script via executing it with --completion)
- [ ] pre-hooks (hooks.d folder will be sourced in order or executed before the real script)
  - [ ] https://umarcor.github.io/cobra/#prerun-and-postrun-hooks
- [x] injects TOME_ROOT into tooling as env var and TOME_EXECUTABLE
- [ ] See if there's utility in ActiveHelp https://umarcor.github.io/cobra/#active-help
- [ ] for hooks we can use the following though post runs won't work alongside our exec pattern
```
	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
```
- [x] Add level and k/v style logging

## Non Features

- Does not support sourcing scripts into shell environment because it adds implementation complexity for other core commands

# Alternate names

kit
edc or eds (everyday scripting)
grimoire
