package git_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/git"
)

// skipIfNoGit skips the test if git is not available in PATH.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH, skipping")
	}
}

// TestRun_ReturnsOutput verifies that a successful command returns its stdout.
func TestRun_ReturnsOutput(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	out, err := git.Run("", "version")
	if err != nil {
		t.Fatalf("git version returned unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "git version") {
		t.Errorf("expected output to begin with 'git version', got %q", out)
	}
}

// TestRun_TrailingNewlineIsTrimmed verifies that trailing newlines are stripped
// from the returned output. Leading whitespace is intentionally preserved so
// that porcelain status lines (e.g. " M file.go") keep their XY column format.
func TestRun_TrailingNewlineIsTrimmed(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	out, err := git.Run("", "version")
	if err != nil {
		t.Fatalf("git version: %v", err)
	}
	if strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\r") {
		t.Errorf("output has trailing newline: %q", out)
	}
}

// TestRun_LeadingWhitespacePreserved verifies that leading whitespace is NOT
// stripped from the output. Git porcelain status lines use position 0 for the
// index column and position 1 for the working-tree column; a space in position
// 0 (e.g. " M file.go") must not be silently removed.
func TestRun_LeadingWhitespacePreserved(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	if _, err := git.Run(dir, "init"); err != nil {
		t.Fatalf("setup: git init: %v", err)
	}
	if _, err := git.Run(dir, "config", "user.email", "test@gyat.test"); err != nil {
		t.Fatalf("setup: git config email: %v", err)
	}
	if _, err := git.Run(dir, "config", "user.name", "gyat test"); err != nil {
		t.Fatalf("setup: git config name: %v", err)
	}
	if _, err := git.Run(dir, "config", "commit.gpgsign", "false"); err != nil {
		t.Fatalf("setup: git config gpgsign: %v", err)
	}

	// Create and commit a file so the repo has a HEAD.
	if err := os.WriteFile(dir+"/init.txt", []byte("init\n"), 0o644); err != nil {
		t.Fatalf("setup: write init.txt: %v", err)
	}
	if _, err := git.Run(dir, "add", "init.txt"); err != nil {
		t.Fatalf("setup: git add: %v", err)
	}
	if _, err := git.Run(dir, "commit", "-m", "initial"); err != nil {
		t.Fatalf("setup: git commit: %v", err)
	}

	// Modify the file without staging it — produces " M init.txt" in porcelain.
	if err := os.WriteFile(dir+"/init.txt", []byte("changed\n"), 0o644); err != nil {
		t.Fatalf("setup: modify init.txt: %v", err)
	}

	out, err := git.Run(dir, "status", "--porcelain")
	if err != nil {
		t.Fatalf("git status --porcelain: %v", err)
	}

	// The first line must start with a space (X=' ', Y='M').
	// If TrimSpace were used, the leading space would be stripped and the
	// porcelain XY columns would be misread by hasWorkingTreeChanges.
	firstLine := strings.SplitN(out, "\n", 2)[0]
	if len(firstLine) < 1 || firstLine[0] != ' ' {
		t.Errorf("expected first porcelain line to start with a space, got %q", firstLine)
	}
}

// TestRun_FailedCommandReturnsError verifies that an invalid git subcommand
// results in a non-nil error.
func TestRun_FailedCommandReturnsError(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	_, err := git.Run("", "not-a-real-git-subcommand")
	if err == nil {
		t.Fatal("expected error for invalid git subcommand, got nil")
	}
}

// TestRun_ErrorContainsStderr verifies that the returned error message contains
// the actual stderr output from git rather than just a bare "exit status N"
// string. This is the key behaviour of our Run wrapper.
func TestRun_ErrorContainsStderr(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	nonExistent := t.TempDir() + "/this_subdir_does_not_exist_xyz"
	_, err := git.Run(nonExistent, "status")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if err.Error() == "exit status 128" {
		t.Errorf(
			"error message is a bare exit status — stderr was not captured: %q",
			err.Error(),
		)
	}
}

// TestRun_EmptyStderrFallsBackToExecError verifies that when git produces no
// stderr output on failure, Run still returns a non-nil error.
func TestRun_EmptyStderrFallsBackToExecError(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	_, err := git.Run("", "version", "--unknown-flag-that-does-not-exist")
	if err == nil {
		t.Skip("git accepted unknown flag, cannot exercise empty-stderr path")
	}
	if err.Error() == "" {
		t.Error("expected a non-empty error message, got an empty string")
	}
}

// TestRun_InitInTempDir verifies that Run can perform a real, stateful git
// operation (init) that creates files on disk.
func TestRun_InitInTempDir(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	out, err := git.Run(dir, "init")
	if err != nil {
		t.Fatalf("git init in temp dir: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output from git init, got empty string")
	}
}

// TestRun_MultipleArgs verifies that multiple arguments are forwarded correctly
// by running `git config --list` which requires at least two tokens.
func TestRun_MultipleArgs(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	if _, err := git.Run(dir, "init"); err != nil {
		t.Fatalf("setup: git init: %v", err)
	}

	_, err := git.Run(dir, "config", "--list")
	if err != nil {
		t.Fatalf("git config --list: %v", err)
	}
}
