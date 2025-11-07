package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// executeWrapperContent executes wrapper script content using bash -c
func executeWrapperContent(content string) ([]byte, error) {
	cmd := exec.Command("bash", "-c", content)
	return cmd.CombinedOutput()
}

// executeWrapperContentWithSh executes wrapper script content using sh -c
func executeWrapperContentWithSh(content string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", content)
	return cmd.CombinedOutput()
}

// TestHookExecution tests end-to-end hook execution
func TestHookExecution(t *testing.T) {
	t.Run("executable hook runs successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a hook that writes to a file
		outputFile := filepath.Join(tmpDir, "hook-output.txt")
		hookContent := `#!/bin/bash
echo "hook executed" > ` + outputFile + `
exit 0
`
		hookPath := filepath.Join(hooksDir, "00-test")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script
		scriptContent := `#!/bin/bash
echo "script executed"
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute wrapper
		output, err := executeWrapperContent(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify hook executed
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Hook did not execute - output file not created")
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(content), "hook executed") {
			t.Errorf("Hook output incorrect: %s", content)
		}
	})

	t.Run("sourced hook modifies environment", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook that exports variable
		hookContent := `export TEST_VAR="from_hook"
export TEST_VAR2="another_value"
`
		hookPath := filepath.Join(hooksDir, "05-env.source")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create script that uses the variable
		outputFile := filepath.Join(tmpDir, "script-output.txt")
		scriptContent := `#!/bin/bash
echo "TEST_VAR=$TEST_VAR" > ` + outputFile + `
echo "TEST_VAR2=$TEST_VAR2" >> ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute wrapper
		output, err := executeWrapperContent(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify script received the environment variable
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		outputStr := string(content)
		if !strings.Contains(outputStr, "TEST_VAR=from_hook") {
			t.Errorf("Sourced hook did not set TEST_VAR correctly: %s", outputStr)
		}
		if !strings.Contains(outputStr, "TEST_VAR2=another_value") {
			t.Errorf("Sourced hook did not set TEST_VAR2 correctly: %s", outputStr)
		}
	})

	t.Run("hook failure aborts execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create failing hook
		hookContent := `#!/bin/bash
echo "hook failed" >&2
exit 1
`
		hookPath := filepath.Join(hooksDir, "00-fail")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Create script that should not execute
		outputFile := filepath.Join(tmpDir, "script-output.txt")
		scriptContent := `#!/bin/bash
echo "script executed" > ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute wrapper - should fail
		output, err := executeWrapperContent(wrapperContent)
		if err == nil {
			t.Fatal("Expected wrapper to fail due to hook failure")
		}

		// Verify error message mentions hook
		if !strings.Contains(string(output), "pre-hook failed") {
			t.Errorf("Error output should mention pre-hook failure: %s", output)
		}

		// Verify script did not execute
		if _, err := os.Stat(outputFile); err == nil {
			t.Error("Script should not have executed after hook failure")
		}
	})

	t.Run("multiple hooks execute in order", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(tmpDir, "execution-order.txt")

		// Create three hooks that append to a file
		hooks := []struct {
			name    string
			content string
		}{
			{"00-first", `#!/bin/bash
echo "first" >> ` + outputFile + `
`},
			{"10-second", `#!/bin/bash
echo "second" >> ` + outputFile + `
`},
			{"20-third", `#!/bin/bash
echo "third" >> ` + outputFile + `
`},
		}

		for _, hook := range hooks {
			hookPath := filepath.Join(hooksDir, hook.name)
			if err := os.WriteFile(hookPath, []byte(hook.content), 0755); err != nil {
				t.Fatal(err)
			}
		}

		// Create target script
		scriptContent := `#!/bin/bash
echo "script" >> ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		discoveredHooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(discoveredHooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}
		// Execute wrapper
		output, err := executeWrapperContent(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify execution order
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		expected := []string{"first", "second", "third", "script"}

		if len(lines) != len(expected) {
			t.Fatalf("Expected %d lines, got %d: %v", len(expected), len(lines), lines)
		}

		for i, line := range lines {
			if line != expected[i] {
				t.Errorf("Line %d: expected '%s', got '%s'", i, expected[i], line)
			}
		}
	})

	t.Run("mixed executable and sourced hooks", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(tmpDir, "mixed-output.txt")

		// Create executable hook
		execHook := `#!/bin/bash
echo "exec hook" >> ` + outputFile + `
`
		execPath := filepath.Join(hooksDir, "00-exec")
		if err := os.WriteFile(execPath, []byte(execHook), 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook
		sourceHook := `export FROM_SOURCE="sourced_value"
echo "source hook" >> ` + outputFile + `
`
		sourcePath := filepath.Join(hooksDir, "05-source.source")
		if err := os.WriteFile(sourcePath, []byte(sourceHook), 0644); err != nil {
			t.Fatal(err)
		}

		// Create script that uses sourced variable
		scriptContent := `#!/bin/bash
echo "script: FROM_SOURCE=$FROM_SOURCE" >> ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute wrapper
		output, err := executeWrapperContent(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify output
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		outputStr := string(content)
		if !strings.Contains(outputStr, "exec hook") {
			t.Error("Executable hook did not execute")
		}
		if !strings.Contains(outputStr, "source hook") {
			t.Error("Sourced hook did not execute")
		}
		if !strings.Contains(outputStr, "FROM_SOURCE=sourced_value") {
			t.Errorf("Script did not receive sourced variable: %s", outputStr)
		}
	})

	t.Run("hook receives environment variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(tmpDir, "env-output.txt")

		// Create hook that writes env vars to file
		hookContent := `#!/bin/bash
echo "TOME_ROOT=$TOME_ROOT" > ` + outputFile + `
echo "TOME_EXECUTABLE=$TOME_EXECUTABLE" >> ` + outputFile + `
echo "TOME_SCRIPT_PATH=$TOME_SCRIPT_PATH" >> ` + outputFile + `
echo "TOME_SCRIPT_NAME=$TOME_SCRIPT_NAME" >> ` + outputFile + `
echo "TOME_SCRIPT_ARGS=$TOME_SCRIPT_ARGS" >> ` + outputFile + `
`
		hookPath := filepath.Join(hooksDir, "00-env-check")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script
		scriptPath := filepath.Join(tmpDir, "my-script")
		scriptContent := `#!/bin/bash
exit 0
`
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "test-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{"arg1", "arg2"})
		if err != nil {
			t.Fatal(err)
		}
		// Execute wrapper
		output, err := executeWrapperContent(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify environment variables
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		outputStr := string(content)
		checks := map[string]string{
			"TOME_ROOT":        tmpDir,
			"TOME_EXECUTABLE":  "test-cli",
			"TOME_SCRIPT_PATH": scriptPath,
			"TOME_SCRIPT_NAME": "my-script",
			"TOME_SCRIPT_ARGS": "arg1 arg2",
		}

		for key, expected := range checks {
			expectedLine := key + "=" + expected
			if !strings.Contains(outputStr, expectedLine) {
				t.Errorf("Missing or incorrect %s: expected '%s' in:\n%s", key, expectedLine, outputStr)
			}
		}
	})

	t.Run("sourced hook failure aborts execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook that fails
		hookContent := `set -e
echo "This will fail" >&2
false
`
		hookPath := filepath.Join(hooksDir, "05-fail.source")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create script that should not execute
		outputFile := filepath.Join(tmpDir, "script-output.txt")
		scriptContent := `#!/bin/bash
echo "script executed" > ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute wrapper - should fail
		_, err = executeWrapperContent(wrapperContent)
		if err == nil {
			t.Fatal("Expected wrapper to fail due to sourced hook failure")
		}

		// Verify script did not execute
		if _, err := os.Stat(outputFile); err == nil {
			t.Error("Script should not have executed after sourced hook failure")
		}
	})
}

// TestShellCompatibility tests that wrapper scripts work with sh, not just bash
func TestShellCompatibility(t *testing.T) {
	t.Run("wrapper script executes with sh", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a simple POSIX-compliant hook
		outputFile := filepath.Join(tmpDir, "hook-output.txt")
		hookContent := `#!/bin/sh
echo "hook executed" > ` + outputFile + `
exit 0
`
		hookPath := filepath.Join(hooksDir, "00-test-hook")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script
		scriptOutput := filepath.Join(tmpDir, "script-output.txt")
		scriptContent := `#!/bin/sh
echo "script executed" > ` + scriptOutput + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Generate wrapper and execute with sh
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)
		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute with sh explicitly
		output, err := executeWrapperContentWithSh(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution with sh failed: %v, output: %s", err, output)
		}

		// Verify hook executed
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Hook did not execute")
		}

		// Verify script executed
		if _, err := os.Stat(scriptOutput); os.IsNotExist(err) {
			t.Error("Script did not execute")
		}
	})

	t.Run("sourced hook works with sh", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create sourced hook with POSIX syntax
		hookContent := `# POSIX-compliant environment setup
TEST_VAR="from_hook"
export TEST_VAR
`
		hookPath := filepath.Join(hooksDir, "05-env.source")
		if err := os.WriteFile(hookPath, []byte(hookContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create script that checks the variable
		outputFile := filepath.Join(tmpDir, "output.txt")
		scriptContent := `#!/bin/sh
echo "TEST_VAR=$TEST_VAR" > ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "test-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Generate and execute with sh
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)
		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		output, err := executeWrapperContentWithSh(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution with sh failed: %v, output: %s", err, output)
		}

		// Verify variable was set
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(content), "TEST_VAR=from_hook") {
			t.Errorf("Variable not set correctly. Got: %s", content)
		}
	})

	t.Run("wrapper uses POSIX-compliant syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		// Create minimal hooks with tmpDir-based path
		hookPath := filepath.Join(tmpDir, "hook1")
		hooks := []Hook{
			{
				Path:    hookPath,
				Name:    "00-hook1",
				Sourced: false,
			},
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, "/fake/script", []string{"arg1", "arg2"})
		if err != nil {
			t.Fatal(err)
		}

		// Verify wrapper doesn't use bash-specific features
		// Check it uses 'set -e' which is POSIX
		if !strings.Contains(wrapperContent, "set -e") {
			t.Error("Wrapper should use POSIX 'set -e'")
		}

		// Verify it doesn't use bash-specific features like [[
		if strings.Contains(wrapperContent, "[[") {
			t.Error("Wrapper should not use bash-specific [[ syntax")
		}

		// Verify exec command is POSIX
		if !strings.Contains(wrapperContent, "exec /fake/script") {
			t.Error("Wrapper should use POSIX exec")
		}
	})
}

// TestHookScenarios tests realistic hook scenarios
func TestHookScenarios(t *testing.T) {
	t.Run("scenario: environment validation and setup", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Hook 1: Validate required variable exists
		validateHook := `#!/bin/bash
if [ -z "$REQUIRED_VAR" ]; then
  echo "Error: REQUIRED_VAR not set" >&2
  exit 1
fi
echo "Validation passed"
`
		validatePath := filepath.Join(hooksDir, "00-validate")
		if err := os.WriteFile(validatePath, []byte(validateHook), 0755); err != nil {
			t.Fatal(err)
		}

		// Hook 2: Set additional environment variables
		setupHook := `export AWS_REGION="us-east-1"
export LOG_LEVEL="info"
`
		setupPath := filepath.Join(hooksDir, "05-setup.source")
		if err := os.WriteFile(setupPath, []byte(setupHook), 0644); err != nil {
			t.Fatal(err)
		}

		// Script that uses the environment
		outputFile := filepath.Join(tmpDir, "script-env.txt")
		scriptContent := `#!/bin/bash
echo "AWS_REGION=$AWS_REGION" > ` + outputFile + `
echo "LOG_LEVEL=$LOG_LEVEL" >> ` + outputFile + `
echo "REQUIRED_VAR=$REQUIRED_VAR" >> ` + outputFile + `
exit 0
`
		scriptPath := filepath.Join(tmpDir, "deploy-script")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// Execute with REQUIRED_VAR set
		cmd := exec.Command("bash", "-c", wrapperContent)
		cmd.Env = append(os.Environ(), "REQUIRED_VAR=test_value")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify script received all environment variables
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		outputStr := string(content)
		if !strings.Contains(outputStr, "AWS_REGION=us-east-1") {
			t.Error("AWS_REGION not set correctly")
		}
		if !strings.Contains(outputStr, "LOG_LEVEL=info") {
			t.Error("LOG_LEVEL not set correctly")
		}
		if !strings.Contains(outputStr, "REQUIRED_VAR=test_value") {
			t.Error("REQUIRED_VAR not passed through")
		}
	})

	t.Run("scenario: audit logging", func(t *testing.T) {
		tmpDir := t.TempDir()
		hooksDir := filepath.Join(tmpDir, ".hooks.d")
		if err := os.Mkdir(hooksDir, 0755); err != nil {
			t.Fatal(err)
		}

		auditLog := filepath.Join(tmpDir, "audit.log")

		// Create audit logging hook
		auditHook := `#!/bin/bash
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "$timestamp | $TOME_SCRIPT_NAME | $TOME_SCRIPT_ARGS" >> ` + auditLog + `
`
		auditPath := filepath.Join(hooksDir, "20-audit")
		if err := os.WriteFile(auditPath, []byte(auditHook), 0755); err != nil {
			t.Fatal(err)
		}

		// Create target script
		scriptContent := `#!/bin/bash
echo "Doing work..."
exit 0
`
		scriptPath := filepath.Join(tmpDir, "sensitive-operation")
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		config := setupTestConfig(t, tmpDir, "tome-cli")
		hr := NewHookRunner(config)

		hooks, err := hr.DiscoverHooks()
		if err != nil {
			t.Fatal(err)
		}

		wrapperContent, err := hr.GenerateWrapperScriptContent(hooks, scriptPath, []string{"--env", "production"})
		if err != nil {
			t.Fatal(err)
		}
		// Execute wrapper
		output, err := executeWrapperContent(wrapperContent)
		if err != nil {
			t.Fatalf("Wrapper execution failed: %v, output: %s", err, output)
		}

		// Verify audit log was created
		content, err := os.ReadFile(auditLog)
		if err != nil {
			t.Fatal(err)
		}

		logStr := string(content)
		if !strings.Contains(logStr, "sensitive-operation") {
			t.Error("Audit log missing script name")
		}
		if !strings.Contains(logStr, "--env production") {
			t.Error("Audit log missing script arguments")
		}
	})
}
