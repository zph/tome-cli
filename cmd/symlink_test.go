package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	gitignore "github.com/sabhiram/go-gitignore"
)

// SYMLINK-001: WHEN a symlinked executable file exists in the root directory,
// the help command SHALL list it alongside regular executable files.
func TestCollectExecutables_IncludesSymlinkedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir, "tome-cli")

	// Create a real executable script
	realScript := filepath.Join(tmpDir, "real-script")
	if err := os.WriteFile(realScript, []byte("#!/bin/bash\n# USAGE: $0 <arg>\necho 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink to it
	symlink := filepath.Join(tmpDir, "linked-script")
	if err := os.Symlink("real-script", symlink); err != nil {
		t.Fatal(err)
	}

	ignorePatterns := gitignore.CompileIgnoreLines()
	executables, err := collectExecutables(tmpDir, ignorePatterns)
	if err != nil {
		t.Fatalf("collectExecutables() returned error: %v", err)
	}

	// Both real and symlinked scripts should appear
	if len(executables) != 2 {
		t.Fatalf("expected 2 executables, got %d: %v", len(executables), executables)
	}

	sort.Strings(executables)
	expected := []string{
		filepath.Join(tmpDir, "linked-script"),
		filepath.Join(tmpDir, "real-script"),
	}
	sort.Strings(expected)

	for i, e := range expected {
		if executables[i] != e {
			t.Errorf("expected executables[%d]=%s, got %s", i, e, executables[i])
		}
	}
}

// SYMLINK-002: WHEN a symlinked executable file exists in a subdirectory,
// the help command SHALL list it alongside regular executable files in that subdirectory.
func TestCollectExecutables_IncludesSymlinksInSubdirs(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir, "tome-cli")

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a real executable in subdir
	realScript := filepath.Join(subDir, "real-script")
	if err := os.WriteFile(realScript, []byte("#!/bin/bash\necho 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink in subdir
	symlink := filepath.Join(subDir, "linked-script")
	if err := os.Symlink("real-script", symlink); err != nil {
		t.Fatal(err)
	}

	ignorePatterns := gitignore.CompileIgnoreLines()
	executables, err := collectExecutables(tmpDir, ignorePatterns)
	if err != nil {
		t.Fatalf("collectExecutables() returned error: %v", err)
	}

	if len(executables) != 2 {
		t.Fatalf("expected 2 executables, got %d: %v", len(executables), executables)
	}

	sort.Strings(executables)
	expected := []string{
		filepath.Join(subDir, "linked-script"),
		filepath.Join(subDir, "real-script"),
	}
	sort.Strings(expected)

	for i, e := range expected {
		if executables[i] != e {
			t.Errorf("expected executables[%d]=%s, got %s", i, e, executables[i])
		}
	}
}

// SYMLINK-003: WHEN a symlink points to a non-existent target (broken symlink),
// the help command SHALL skip it without error.
func TestCollectExecutables_SkipsBrokenSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir, "tome-cli")

	// Create a real executable
	realScript := filepath.Join(tmpDir, "real-script")
	if err := os.WriteFile(realScript, []byte("#!/bin/bash\necho 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a broken symlink (target does not exist)
	brokenLink := filepath.Join(tmpDir, "broken-link")
	if err := os.Symlink("nonexistent-target", brokenLink); err != nil {
		t.Fatal(err)
	}

	ignorePatterns := gitignore.CompileIgnoreLines()
	executables, err := collectExecutables(tmpDir, ignorePatterns)
	if err != nil {
		t.Fatalf("collectExecutables() returned error: %v", err)
	}

	// Only the real script should appear, broken symlink skipped
	if len(executables) != 1 {
		t.Fatalf("expected 1 executable, got %d: %v", len(executables), executables)
	}
	if executables[0] != realScript {
		t.Errorf("expected %s, got %s", realScript, executables[0])
	}
}

// SYMLINK-005: WHEN a symlinked executable file exists in the root directory,
// the help command with the script name as argument SHALL display its usage and help text.
func TestNewScript_FollowsSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir, "tome-cli")

	// Create a real script with USAGE header
	realScript := filepath.Join(tmpDir, "real-script")
	content := "#!/bin/bash\n# USAGE: $0 <arg1> <arg2>\n# This is help text\n\necho 1\n"
	if err := os.WriteFile(realScript, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	symlink := filepath.Join(tmpDir, "linked-script")
	if err := os.Symlink("real-script", symlink); err != nil {
		t.Fatal(err)
	}

	// NewScript should parse the symlink target and extract usage/help
	s := NewScript(symlink, tmpDir)
	usage := s.Usage()
	if usage == "" {
		t.Error("NewScript on symlink returned empty usage, expected parsed usage from target")
	}
	if usage != "<arg1> <arg2>" {
		t.Errorf("expected usage '<arg1> <arg2>', got '%s'", usage)
	}

	help := s.Help()
	if help == "" {
		t.Error("NewScript on symlink returned empty help, expected parsed help from target")
	}
}

// SYMLINK-006: WHEN a symlinked executable file exists in the root directory,
// the completion system SHALL include it in tab completion results.
func TestCompletionIncludesSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestConfig(t, tmpDir, "tome-cli")

	// Create a real executable script
	realScript := filepath.Join(tmpDir, "real-script")
	if err := os.WriteFile(realScript, []byte("#!/bin/bash\n# USAGE: $0 <arg>\necho 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	symlink := filepath.Join(tmpDir, "linked-script")
	if err := os.Symlink("real-script", symlink); err != nil {
		t.Fatal(err)
	}

	// ReadDir is used by ValidArgsFunctionForScripts; verify symlink appears
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir() returned error: %v", err)
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	foundLinked := false
	foundReal := false
	for _, name := range names {
		if name == "linked-script" {
			foundLinked = true
		}
		if name == "real-script" {
			foundReal = true
		}
	}

	if !foundLinked {
		t.Error("linked-script not found in directory listing for completion")
	}
	if !foundReal {
		t.Error("real-script not found in directory listing for completion")
	}

	// Verify the symlinked script is detected as executable via os.Stat (which follows symlinks)
	info, err := os.Stat(symlink)
	if err != nil {
		t.Fatalf("os.Stat on symlink failed: %v", err)
	}
	if !isExecutableByOwner(info.Mode()) {
		t.Error("symlinked script not detected as executable via os.Stat")
	}
}
