package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// setupTestConfig sets up viper config for testing
func setupTestConfig(t *testing.T, rootDir, executableName string) *Config {
	t.Helper()

	// Initialize logger if not already done
	if log == nil {
		log = createLogger("test", os.Stderr)
	}

	// Save current values
	oldRoot := viper.Get("root")
	oldExec := viper.Get("executable")

	// Set test values
	viper.Set("root", rootDir)
	viper.Set("executable", executableName)

	// Restore on cleanup
	t.Cleanup(func() {
		if oldRoot != nil {
			viper.Set("root", oldRoot)
		}
		if oldExec != nil {
			viper.Set("executable", oldExec)
		}
	})

	return NewConfig()
}

// TestHookDiscovery tests the DiscoverHooks function
func TestHookDiscovery(t *testing.T) {
	t.Run("missing hooks directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error for missing directory: %v", err)
		}
		if len(hooks) != 0 {
			t.Errorf("DiscoverHooks() returned hooks for missing directory, expected 0, got %d", len(hooks))
		}
	})

	t.Run("empty hooks directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(hooks) != 0 {
			t.Errorf("DiscoverHooks() expected 0 hooks, got %d", len(hooks))
		}
	})

	t.Run("discovers executable hooks", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create executable hook
		hookPath := filepath.Join(hooksDir, "00-test-hook")
		if err := os.WriteFile(hookPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(hooks) != 1 {
			t.Fatalf("DiscoverHooks() expected 1 hook, got %d", len(hooks))
		}
		if hooks[0].Name != "00-test-hook" {
			t.Errorf("Hook name expected '00-test-hook', got '%s'", hooks[0].Name)
		}
		if hooks[0].Sourced {
			t.Errorf("Hook should not be marked as sourced")
		}
	})

	t.Run("discovers sourced hooks with .source suffix", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook (no execute permission needed)
		hookPath := filepath.Join(hooksDir, "05-env.source")
		if err := os.WriteFile(hookPath, []byte("export FOO=bar"), 0644); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(hooks) != 1 {
			t.Fatalf("DiscoverHooks() expected 1 hook, got %d", len(hooks))
		}
		if hooks[0].Name != "05-env.source" {
			t.Errorf("Hook name expected '05-env.source', got '%s'", hooks[0].Name)
		}
		if !hooks[0].Sourced {
			t.Errorf("Hook should be marked as sourced")
		}
	})

	t.Run("skips non-executable files without .source suffix", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create non-executable hook without .source suffix
		hookPath := filepath.Join(hooksDir, "10-no-exec")
		if err := os.WriteFile(hookPath, []byte("#!/bin/bash\necho test"), 0644); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(hooks) != 0 {
			t.Errorf("DiscoverHooks() should skip non-executable hook, got %d hooks", len(hooks))
		}
	})

	t.Run("sorts hooks lexicographically", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create hooks in random order
		hooks := []string{"20-third", "00-first", "10-second"}
		for _, hook := range hooks {
			hookPath := filepath.Join(hooksDir, hook)
			if err := os.WriteFile(hookPath, []byte("#!/bin/bash\n"), 0755); err != nil {
				t.Fatal(err)
			}
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		discoveredHooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(discoveredHooks) != 3 {
			t.Fatalf("DiscoverHooks() expected 3 hooks, got %d", len(discoveredHooks))
		}

		expected := []string{"00-first", "10-second", "20-third"}
		for i, hook := range discoveredHooks {
			if hook.Name != expected[i] {
				t.Errorf("Hook at index %d expected '%s', got '%s'", i, expected[i], hook.Name)
			}
		}
	})

	t.Run("skips directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a subdirectory
		subDir := filepath.Join(hooksDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(hooks) != 0 {
			t.Errorf("DiscoverHooks() should skip directories, got %d hooks", len(hooks))
		}
	})

	t.Run("mixed executable and sourced hooks", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create executable hook
		execHook := filepath.Join(hooksDir, "00-exec")
		if err := os.WriteFile(execHook, []byte("#!/bin/bash\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook
		sourceHook := filepath.Join(hooksDir, "05-env.source")
		if err := os.WriteFile(sourceHook, []byte("export FOO=bar"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create another executable hook
		execHook2 := filepath.Join(hooksDir, "10-another")
		if err := os.WriteFile(execHook2, []byte("#!/bin/bash\n"), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Errorf("DiscoverHooks() returned error: %v", err)
		}
		if len(hooks) != 3 {
			t.Fatalf("DiscoverHooks() expected 3 hooks, got %d", len(hooks))
		}

		// Verify order and types
		if !hooks[0].Sourced && hooks[0].Name != "00-exec" {
			t.Errorf("First hook should be executable '00-exec'")
		}
		if hooks[1].Sourced && hooks[1].Name != "05-env.source" {
			t.Errorf("Second hook should be sourced '05-env.source'")
		}
		if !hooks[2].Sourced && hooks[2].Name != "10-another" {
			t.Errorf("Third hook should be executable '10-another'")
		}
	})
}

// TestGenerateWrapperScriptContent tests the wrapper script generation
func TestGenerateWrapperScriptContent(t *testing.T) {
	t.Run("no hooks returns empty content", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		content, err := hr.GenerateWrapperScriptContent([]Hook{}, "/fake/script", []string{})
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content != "" {
			t.Errorf("GenerateWrapperScriptContent() expected empty content for no hooks, got '%s'", content)
		}
	})

	t.Run("generates wrapper with executable hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hookPath := filepath.Join(tmpDir, "00-test")
		hooks := []Hook{
			{
				Path:    hookPath,
				Name:    "00-test",
				Sourced: false,
			},
		}

		content, err := hr.GenerateWrapperScriptContent(hooks, "/fake/script", []string{"arg1", "arg2"})
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content == "" {
			t.Fatal("GenerateWrapperScriptContent() returned empty content")
		}

		// Check for hook execution (now quoted)
		if !contains(content, `"`+hookPath+`"`) {
			t.Error("Wrapper script missing hook execution")
		}
		// Check for target script exec (shellescape only quotes when necessary)
		if !contains(content, `exec "/fake/script" arg1 arg2`) {
			t.Error("Wrapper script missing target script exec")
		}
	})

	t.Run("generates wrapper with sourced hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hookPath := filepath.Join(tmpDir, "05-env.source")
		hooks := []Hook{
			{
				Path:    hookPath,
				Name:    "05-env.source",
				Sourced: true,
			},
		}

		content, err := hr.GenerateWrapperScriptContent(hooks, "/fake/script", []string{})
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content == "" {
			t.Fatal("GenerateWrapperScriptContent() returned empty content")
		}

		contentStr := content
		// Check for source command (now quoted)
		if !contains(contentStr, `source "`+hookPath+`"`) {
			t.Error("Wrapper script missing source command for .source hook")
		}
	})

	t.Run("wrapper includes environment variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hookPath := filepath.Join(tmpDir, "00-test")
		hooks := []Hook{
			{
				Path:    hookPath,
				Name:    "00-test",
				Sourced: false,
			},
		}

		content, err := hr.GenerateWrapperScriptContent(hooks, "/fake/script", []string{"arg1"})
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content == "" {
			t.Fatal("GenerateWrapperScriptContent() returned empty content")
		}

		contentStr := content
		// Check for environment variables
		if !contains(contentStr, "TOME_ROOT=") {
			t.Error("Wrapper script missing TOME_ROOT")
		}
		if !contains(contentStr, "TOME_EXECUTABLE=") {
			t.Error("Wrapper script missing TOME_EXECUTABLE")
		}
		if !contains(contentStr, "TOME_SCRIPT_PATH=") {
			t.Error("Wrapper script missing TOME_SCRIPT_PATH")
		}
		if !contains(contentStr, "TOME_SCRIPT_NAME=") {
			t.Error("Wrapper script missing TOME_SCRIPT_NAME")
		}
		if !contains(contentStr, "TOME_SCRIPT_ARGS=") {
			t.Error("Wrapper script missing TOME_SCRIPT_ARGS")
		}
	})

	t.Run("wrapper is executable", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hookPath := filepath.Join(tmpDir, "00-test")
		hooks := []Hook{
			{
				Path:    hookPath,
				Name:    "00-test",
				Sourced: false,
			},
		}

		content, err := hr.GenerateWrapperScriptContent(hooks, "/fake/script", []string{})
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content == "" {
			t.Fatal("GenerateWrapperScriptContent() returned empty content")
		}

		// Verify content is valid shell script with expected structure
		if !contains(content, "set -e") {
			t.Error("Wrapper script missing 'set -e' directive")
		}
		if !contains(content, `exec "/fake/script"`) {
			t.Error("Wrapper script missing exec command for target script")
		}
	})

	t.Run("wrapper handles multiple hooks in order", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hook1Path := filepath.Join(tmpDir, "00-first")
		hook2Path := filepath.Join(tmpDir, "05-env.source")
		hook3Path := filepath.Join(tmpDir, "10-second")

		hooks := []Hook{
			{
				Path:    hook1Path,
				Name:    "00-first",
				Sourced: false,
			},
			{
				Path:    hook2Path,
				Name:    "05-env.source",
				Sourced: true,
			},
			{
				Path:    hook3Path,
				Name:    "10-second",
				Sourced: false,
			},
		}

		content, err := hr.GenerateWrapperScriptContent(hooks, "/fake/script", []string{})
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content == "" {
			t.Fatal("GenerateWrapperScriptContent() returned empty content")
		}

		contentStr := content
		// Verify all hooks are present (now with quotes)
		if !contains(contentStr, `"`+hook1Path+`"`) {
			t.Error("Wrapper script missing first hook")
		}
		if !contains(contentStr, `source "`+hook2Path+`"`) {
			t.Error("Wrapper script missing sourced hook")
		}
		if !contains(contentStr, `"`+hook3Path+`"`) {
			t.Error("Wrapper script missing second hook")
		}

		// Verify order (crude check - first hook should appear before second)
		firstIdx := indexOf(contentStr, "00-first")
		secondIdx := indexOf(contentStr, "05-env.source")
		thirdIdx := indexOf(contentStr, "10-second")

		if firstIdx == -1 || secondIdx == -1 || thirdIdx == -1 {
			t.Error("Not all hooks found in wrapper")
		}
		if firstIdx > secondIdx || secondIdx > thirdIdx {
			t.Error("Hooks not in correct order")
		}
	})

	t.Run("wrapper handles paths with spaces", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		// Create hook and script paths with spaces
		hookPath := filepath.Join(tmpDir, "hook with spaces")
		scriptPath := filepath.Join(tmpDir, "script with spaces.sh")

		hooks := []Hook{
			{
				Path:    hookPath,
				Name:    "hook with spaces",
				Sourced: false,
			},
		}

		args := []string{"arg with spaces", "normal-arg", "another arg"}
		content, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, args)
		if err != nil {
			t.Errorf("GenerateWrapperScriptContent() returned error: %v", err)
		}
		if content == "" {
			t.Fatal("GenerateWrapperScriptContent() returned empty content")
		}

		// Verify hook path is quoted
		if !contains(content, `"`+hookPath+`"`) {
			t.Error("Wrapper script missing quoted hook path")
		}

		// Verify script path is quoted
		if !contains(content, `"`+scriptPath+`"`) {
			t.Error("Wrapper script missing quoted script path")
		}

		// Verify args with spaces are quoted but normal args are not
		if !contains(content, `'arg with spaces'`) {
			t.Error("Wrapper script missing quoted argument with spaces")
		}
		if !contains(content, `normal-arg`) {
			t.Error("Wrapper script missing unquoted simple argument")
		}
		if !contains(content, `'another arg'`) {
			t.Error("Wrapper script missing second quoted argument with spaces")
		}
	})
}

// TestBuildHookEnv tests environment variable building
func TestBuildHookEnv(t *testing.T) {
	t.Run("builds standard environment variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		env := hr.buildHookEnv("", "/path/to/script", []string{"arg1", "arg2"})

		// Check for required variables
		vars := map[string]bool{
			"TOME_ROOT=":         false,
			"TOME_EXECUTABLE=":   false,
			"TOME_SCRIPT_PATH=":  false,
			"TOME_SCRIPT_NAME=":  false,
			"TOME_SCRIPT_ARGS=":  false,
			"TOME_CLI_ROOT=":     false, // uppercase snake case of executable
			"TOME_CLI_EXECUTABLE=": false,
		}

		for _, e := range env {
			for key := range vars {
				if contains(e, key) {
					vars[key] = true
				}
			}
		}

		for key, found := range vars {
			if !found {
				t.Errorf("Missing environment variable: %s", key)
			}
		}
	})

	t.Run("includes script args", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		env := hr.buildHookEnv("", "/path/to/script", []string{"arg1", "arg2", "arg3"})

		found := false
		for _, e := range env {
			// Args are quoted to handle spaces properly in bash
			if contains(e, `TOME_SCRIPT_ARGS="arg1 arg2 arg3"`) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("TOME_SCRIPT_ARGS not set correctly. Got: %v", env)
		}
	})

	t.Run("uses custom executable name", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "my-custom-cli")
		hr := NewHookRunner(config)

		env := hr.buildHookEnv("", "/path/to/script", []string{})

		found := false
		for _, e := range env {
			if contains(e, "MY_CUSTOM_CLI_ROOT=") {
				found = true
				break
			}
		}

		if !found {
			t.Error("Custom executable environment variable not set")
		}
	})
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) != -1
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
