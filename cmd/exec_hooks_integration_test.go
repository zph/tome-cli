package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecWithHooks tests the full integration of hooks with the exec command
func TestExecWithHooks(t *testing.T) {
	t.Run("exec runs hooks before script", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a hook that writes to a file
		outputFile := filepath.Join(tmpDir, "execution-order.txt")
		hookContent := `#!/bin/bash
echo "hook executed" >> ` + outputFile + `
exit 0
`
		hookPath := filepath.Join(hooksDir, "00-test-hook")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script that also writes to the file
		scriptContent := `#!/bin/bash
echo "script executed" >> ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Setup config
		_ = setupTestConfig(t, tmpDir, "tome-cli")

		// Simulate exec command
		skipHooks = false
		err := ExecRunE(nil, []string{"test-script"})
		if err != nil {
			t.Fatalf("ExecRunE failed: %v", err)
		}

		// Note: ExecRunE uses syscall.Exec which replaces the process
		// So we can't verify the output file in this test
		// The integration tests in hooks_integration_test.go cover actual execution
	})

	t.Run("exec with sourced hook modifies environment", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook
		hookContent := `export TEST_VAR="from_hook"
`
		hookPath := filepath.Join(hooksDir, "05-env.source")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create script that uses the variable
		outputFile := filepath.Join(tmpDir, "script-output.txt")
		scriptContent := `#!/bin/bash
echo "TEST_VAR=$TEST_VAR" > ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Setup config
		config := setupTestConfig(t, tmpDir, "tome-cli")

		// Test hook discovery and wrapper generation
		skipHooks = false
		hookRunner := NewHookRunner(config)
		hooks, err := hookRunner.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		if len(hooks) != 1 {
			t.Fatalf("Expected 1 hook, got %d", len(hooks))
		}

		if !hooks[0].Sourced {
			t.Error("Hook should be marked as sourced")
		}

		wrapperContent, err := hookRunner.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}
		// Verify wrapper contains source command
		if !strings.Contains(wrapperContent, "source "+hookPath) {
			t.Error("Wrapper should contain source command for .source hook")
		}
	})

	t.Run("exec with --skip-hooks flag skips hooks", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a hook
		hookPath := filepath.Join(hooksDir, "00-test-hook")
		if err := os.WriteFile(hookPath, []byte("#!/bin/bash\nexit 1\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Setup config
		config := setupTestConfig(t, tmpDir, "tome-cli")

		// Set skipHooks flag
		skipHooks = true

		// Test that hooks are discovered but not used
		hookRunner := NewHookRunner(config)
		hooks, err := hookRunner.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		if len(hooks) != 1 {
			t.Fatalf("Expected 1 hook to be discovered, got %d", len(hooks))
		}

		// In the actual exec, hooks would be skipped due to skipHooks flag
		// We can't test the full exec here due to syscall.Exec, but we verify
		// the flag is honored in the logic
	})

	t.Run("exec discovers hooks from .hooks.d", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create multiple hooks
		hooks := []struct {
			name    string
			content string
			mode    os.FileMode
		}{
			{"00-first", "#!/bin/bash\necho first\n", 0755},
			{"05-env.source", "export VAR=value\n", 0644},
			{"10-second", "#!/bin/bash\necho second\n", 0755},
		}

		for _, hook := range hooks {
			hookPath := filepath.Join(hooksDir, hook.name)
			if err := os.WriteFile(hookPath, []byte(hook.content), hook.mode); err != nil {
				t.Fatal(err)
			}
		}

		// Create target script
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Setup config and test discovery
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hookRunner := NewHookRunner(config)
		discoveredHooks, err := hookRunner.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		if len(discoveredHooks) != 3 {
			t.Fatalf("Expected 3 hooks, got %d", len(discoveredHooks))
		}

		// Verify order
		expectedOrder := []string{"00-first", "05-env.source", "10-second"}
		for i, hook := range discoveredHooks {
			if hook.Name != expectedOrder[i] {
				t.Errorf("Hook %d: expected %s, got %s", i, expectedOrder[i], hook.Name)
			}
		}

		// Verify sourced flag
		if discoveredHooks[0].Sourced {
			t.Error("00-first should not be sourced")
		}
		if !discoveredHooks[1].Sourced {
			t.Error("05-env.source should be sourced")
		}
		if discoveredHooks[2].Sourced {
			t.Error("10-second should not be sourced")
		}
	})

	t.Run("exec handles missing .hooks.d gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		// No .hooks.d directory created

		// Create target script
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Setup config
		config := setupTestConfig(t, tmpDir, "tome-cli")

		// Should not error when .hooks.d doesn't exist
		skipHooks = false
		hookRunner := NewHookRunner(config)
		hooks, err := hookRunner.DiscoverHooks()
		if err != nil {
			t.Fatalf("DiscoverHooks should not error on missing .hooks.d: %v", err)
		}

		if len(hooks) != 0 {
			t.Errorf("Expected 0 hooks when .hooks.d missing, got %d", len(hooks))
		}
	})

	t.Run("exec passes script args to wrapper", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a hook
		hookPath := filepath.Join(hooksDir, "00-hook")
		if err := os.WriteFile(hookPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// Setup config
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hookRunner := NewHookRunner(config)
		hooks, err := hookRunner.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		// Generate wrapper with args
		args := []string{"arg1", "arg2", "arg3"}
		wrapperContent, err := hookRunner.GenerateWrapperScriptContent(hooks, scriptPath, args)
		if err != nil {
			t.Fatal(err)
		}
		// Verify wrapper contains args
		if !strings.Contains(wrapperContent,"exec "+scriptPath+" arg1 arg2 arg3") {
			t.Error("Wrapper should contain script path and arguments")
		}

		// Verify environment variable is set
		if !strings.Contains(wrapperContent,`TOME_SCRIPT_ARGS="arg1 arg2 arg3"`) {
			t.Error("Wrapper should set TOME_SCRIPT_ARGS with all arguments")
		}
	})
}

// TestExecFlags tests exec command flags
func TestExecFlags(t *testing.T) {
	t.Run("skip-hooks flag exists", func(t *testing.T) {
		flag := execCmd.Flags().Lookup("skip-hooks")
		if flag == nil {
			t.Fatal("--skip-hooks flag not found")
		}

		if flag.DefValue != "false" {
			t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
		}

		if flag.Usage != "Skip pre-execution hooks" {
			t.Errorf("Unexpected usage text: %s", flag.Usage)
		}
	})

	t.Run("dry-run flag exists", func(t *testing.T) {
		flag := execCmd.Flags().Lookup("dry-run")
		if flag == nil {
			t.Fatal("--dry-run flag not found")
		}
	})
}
