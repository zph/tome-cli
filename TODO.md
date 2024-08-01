# TODO

- [ ] Update/replace all placeholder text from cobra
- [ ] setup default required flags/env var for root folder
- [ ] decide if the config file is going to be used and how to use it with multiple `tome`
  - don't use config file for now
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
- [ ] root argument in config file respects Env vars with syntax of go templates: `{{ Env 'HOME' }}`
