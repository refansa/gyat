package cmd

import (
	"bytes"
	"io"
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

// TestRunPush_SkipsLocalPathSubmodule verifies that a submodule whose URL is a
// local path is skipped with a hint rather than attempting a remote push.
func TestRunPush_SkipsLocalPathSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-push-local")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPush(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPush with local path submodule: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "hint:") {
		t.Errorf("expected stderr to contain 'hint:', got:\n%s", stderr)
	}
	if !strings.Contains(stderr, subPath) {
		t.Errorf("expected stderr to mention submodule %q, got:\n%s", subPath, stderr)
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

// TestRunPush_SkipsLocalPathEvenAfterCommit verifies that a submodule whose
// URL is a local path is not pushed even when new local commits exist — the
// hint is printed and the commit does not appear in the source repository.
func TestRunPush_SkipsLocalPathEvenAfterCommit(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-push-local-commit")

	// Rename the default branch to "main" for determinism.
	runGitIn(t, source, "branch", "-M", "main")

	// Allow pushes to the non-bare source repository (needed only if push
	// were actually attempted; included here to avoid a false-negative).
	runGitIn(t, source, "config", "receive.denyCurrentBranch", "updateInstead")

	rel := relPath(umbrella, source)
	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "main", tc, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subDir := filepath.Join(umbrella, "svc-push-local-commit")

	// Check out the named branch so the submodule is not in detached HEAD.
	runGitIn(t, subDir, "checkout", "main")

	// Create and stage a new commit inside the submodule.
	writeFile(t, filepath.Join(subDir, "pushed.go"), "package main\n")
	runGitIn(t, subDir, "add", ".")
	runGitIn(t, subDir, "commit", "-m", "feat: local only commit")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPush(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPush: %v", err)
	}

	// The commit must NOT have been pushed to the source repository.
	sourceLog := runGitIn(t, source, "log", "--oneline", "-1")
	if strings.Contains(sourceLog, "feat: local only commit") {
		t.Errorf("expected commit to remain local (not pushed), but it appeared in source: %s", sourceLog)
	}

	// A hint message must have been printed.
	if !strings.Contains(stderrBuf.String(), "hint:") {
		t.Errorf("expected hint message in stderr, got:\n%s", stderrBuf.String())
	}
}
