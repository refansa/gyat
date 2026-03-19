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

// ---------------------------------------------------------------------------
// Unit tests — resolveTargetPaths
// ---------------------------------------------------------------------------

// TestResolveTargetPaths_EmptyArgs verifies that passing no args returns all
// registered submodule paths unchanged.
func TestResolveTargetPaths_EmptyArgs(t *testing.T) {
	t.Parallel()

	paths := []string{"services/auth", "services/billing"}
	got, err := resolveTargetPaths(paths, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(paths) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(paths))
	}
	for i, p := range paths {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

// TestResolveTargetPaths_ValidArgs verifies that args matching registered paths
// are returned normalised with forward slashes.
func TestResolveTargetPaths_ValidArgs(t *testing.T) {
	t.Parallel()

	submodulePaths := []string{"services/auth", "services/billing"}
	args := []string{"services/auth"}

	got, err := resolveTargetPaths(submodulePaths, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0] != "services/auth" {
		t.Errorf("got[0] = %q, want %q", got[0], "services/auth")
	}
}

// TestResolveTargetPaths_InvalidArg verifies that an arg not present in the
// registered submodule paths returns an error containing that arg.
func TestResolveTargetPaths_InvalidArg(t *testing.T) {
	t.Parallel()

	submodulePaths := []string{"services/auth"}
	args := []string{"services/billing"}

	_, err := resolveTargetPaths(submodulePaths, args)
	if err == nil {
		t.Fatal("expected an error for an unregistered path, got nil")
	}
	if !strings.Contains(err.Error(), "services/billing") {
		t.Errorf("expected error to contain %q, got: %v", "services/billing", err)
	}
}

// ---------------------------------------------------------------------------
// Integration tests — runPull
// ---------------------------------------------------------------------------

// TestRunPull_NoSubmodules verifies that pulling in a repo with no submodules
// and no upstream prints "nothing to pull" and returns no error.
func TestRunPull_NoSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPull(dir, false, cc, nil); err != nil {
		t.Fatalf("runPull in empty repo: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "nothing to pull") {
		t.Errorf("expected stderr to contain 'nothing to pull', got:\n%s", stderrBuf.String())
	}
}

// TestRunPull_SkipsUncheckedSubmodule verifies that a submodule whose directory
// has been removed is skipped with a warning rather than causing an error.
func TestRunPull_SkipsUncheckedSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-pull-absent")

	// Remove the submodule directory to simulate an unchecked-out submodule.
	if err := os.RemoveAll(filepath.Join(umbrella, subPath)); err != nil {
		t.Fatalf("removing submodule dir: %v", err)
	}

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPull(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPull with unchecked submodule: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "warning:") {
		t.Errorf("expected stderr to contain 'warning:', got:\n%s", stderr)
	}
	if !strings.Contains(stderr, subPath) {
		t.Errorf("expected stderr to contain submodule name %q, got:\n%s", subPath, stderr)
	}
}

// TestRunPull_InvalidPath verifies that passing a path not registered as a
// submodule returns an error.
func TestRunPull_InvalidPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedSubmodule(t, "svc-pull-bad")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)

	err := runPull(umbrella, false, cc, []string{"nonexistent-sub"})
	if err == nil {
		t.Error("expected an error for an unregistered submodule path, got nil")
	}
}

// TestRunPull_NewCommitIsPickedUp verifies that after a new commit is pushed to
// the source repository, running pull advances the submodule's HEAD.
func TestRunPull_NewCommitIsPickedUp(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-pull-ahead")

	// Rename the default branch to "main" so the tracking branch is consistent.
	runGitIn(t, source, "branch", "-M", "main")

	rel := relPath(umbrella, source)
	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "main", tc, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subDir := filepath.Join(umbrella, "svc-pull-ahead")

	// Check out the named branch so that git pull has a tracking target.
	runGitIn(t, subDir, "checkout", "main")

	before := runGitIn(t, subDir, "rev-parse", "HEAD")

	// Add a new commit to the source repository.
	writeFile(t, filepath.Join(source, "feature.go"), "package main\n// new feature\n")
	runGitIn(t, source, "add", ".")
	runGitIn(t, source, "commit", "-m", "feat: new feature")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)
	if err := runPull(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPull: %v", err)
	}

	after := runGitIn(t, subDir, "rev-parse", "HEAD")

	if before == after {
		t.Errorf("expected submodule HEAD to advance after pull, but it stayed at %s", before)
	}
}
