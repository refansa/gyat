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
// Integration tests — runPush
// ---------------------------------------------------------------------------

// TestRunPush_NoSubmodules verifies that pushing in a repo with no submodules
// and no remote prints "nothing to push" and returns no error.
func TestRunPush_NoSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPush(dir, false, cc, nil); err != nil {
		t.Fatalf("runPush in empty repo: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "nothing to push") {
		t.Errorf("expected stderr to contain 'nothing to push', got:\n%s", stderrBuf.String())
	}
}

// TestRunPush_SkipsUncheckedSubmodule verifies that a submodule whose directory
// has been removed is skipped with a warning rather than causing an error.
func TestRunPush_SkipsUncheckedSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-push-absent")

	// Remove the submodule directory to simulate an unchecked-out submodule.
	if err := os.RemoveAll(filepath.Join(umbrella, subPath)); err != nil {
		t.Fatalf("removing submodule dir: %v", err)
	}

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPush(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPush with unchecked submodule: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "warning:") {
		t.Errorf("expected stderr to contain 'warning:', got:\n%s", stderr)
	}
	if !strings.Contains(stderr, subPath) {
		t.Errorf("expected stderr to contain submodule name %q, got:\n%s", subPath, stderr)
	}
}

// TestRunPush_InvalidPath verifies that passing a path not registered as a
// submodule returns an error.
func TestRunPush_InvalidPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedSubmodule(t, "svc-push-bad")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)

	err := runPush(umbrella, false, cc, []string{"nonexistent-sub"})
	if err == nil {
		t.Error("expected an error for an unregistered submodule path, got nil")
	}
}

// TestRunPush_PushesToRemote verifies that a commit made inside a submodule
// is pushed back to the source repository.
func TestRunPush_PushesToRemote(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-push-remote")

	// Rename the default branch to "main" for consistency.
	runGitIn(t, source, "branch", "-M", "main")

	// Allow pushes to the non-bare source repository.
	runGitIn(t, source, "config", "receive.denyCurrentBranch", "updateInstead")

	rel := relPath(umbrella, source)
	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "main", tc, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subDir := filepath.Join(umbrella, "svc-push-remote")

	// Check out the named branch so that git push has a tracking target.
	runGitIn(t, subDir, "checkout", "main")

	// Create and commit a new file inside the submodule.
	writeFile(t, filepath.Join(subDir, "pushed.go"), "package main\n")
	runGitIn(t, subDir, "add", ".")
	runGitIn(t, subDir, "commit", "-m", "feat: pushed file")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)
	if err := runPush(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPush: %v", err)
	}

	// Verify the commit arrived in the source repository.
	log := runGitIn(t, source, "log", "--oneline", "-1")
	if !strings.Contains(log, "feat: pushed file") {
		t.Errorf("expected source log to contain 'feat: pushed file', got: %s", log)
	}
}
