package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunInit_EmptyDirectory verifies that gyat init creates a .git directory
// in a fresh, empty directory.
func TestRunInit_EmptyDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)

	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit in empty directory: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, ".git"))
}

// TestRunInit_AlreadyAGitRepo verifies that running gyat init in an existing
// git repository is idempotent — it should succeed without error.
func TestRunInit_AlreadyAGitRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)

	// Running init a second time should reinitialise cleanly.
	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit on existing repo returned unexpected error: %v", err)
	}
}

// TestRunInit_PrintsReadyMessage verifies that when there is no .gitmodules
// file, the output tells the user to run gyat track.
func TestRunInit_PrintsReadyMessage(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	var stderrBuf bytes.Buffer
	ic := &cobra.Command{}
	ic.SetErr(&stderrBuf)

	if err := runInit(dir, ic, nil); err != nil {
		t.Errorf("runInit: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "gyat track") {
		t.Errorf("expected hint to mention 'gyat track' on stderr, got:\n%s", stderrBuf.String())
	}
}

// TestRunInit_WithGitmodules verifies that when a .gitmodules file is already
// present in the directory (e.g. after a shallow clone of an umbrella repo),
// gyat init attempts to initialise the submodules.
func TestRunInit_WithGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	// Build a full setup: umbrella with one committed submodule, then clone it.
	umbrella, source := newTestSetup(t, "service-auth")

	// Track and commit the submodule in the umbrella repo.
	rel := relPath(umbrella, source)
	addC := &cobra.Command{}
	addC.SetErr(io.Discard)
	if err := runTrack(umbrella, "", addC, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}
	runGitIn(t, umbrella, "commit", "-m", "add submodule")

	// Clone the umbrella without recursing into submodules.
	cloneBase := t.TempDir()
	runGitIn(t, cloneBase, "clone", umbrella, "cloned")
	cloned := filepath.Join(cloneBase, "cloned")

	// The submodule directory should exist but be empty before init.
	subDir := filepath.Join(cloned, "service-auth")
	entries, err := os.ReadDir(subDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read submodule dir before init: %v", err)
	}
	if len(entries) > 0 {
		t.Skip("submodule was already populated — clone behaviour differs on this git version")
	}

	// Run gyat init inside the clone — it should detect .gitmodules and
	// initialise the submodules automatically.
	var stderrBuf bytes.Buffer
	ic := &cobra.Command{}
	ic.SetErr(&stderrBuf)
	if err := runInit(cloned, ic, nil); err != nil {
		t.Errorf("runInit with .gitmodules: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), ".gitmodules") {
		t.Errorf("expected stderr to mention .gitmodules, got:\n%s", stderrBuf.String())
	}
}

// TestRunInit_DotGitIsDirectory verifies that the .git entry created by init
// is a directory (not a file, which would indicate a git worktree).
func TestRunInit_DotGitIsDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)

	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatalf("stat .git: %v", err)
	}
	if !info.IsDir() {
		t.Error(".git should be a directory for a standard (non-worktree) repository")
	}
}
