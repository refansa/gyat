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
// Unit tests — submoduleURLMap
// ---------------------------------------------------------------------------

// TestSubmoduleURLMap_NoGitmodules verifies that a repository with no
// .gitmodules returns a nil map without an error.
func TestSubmoduleURLMap_NoGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	urlMap, err := submoduleURLMap(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if urlMap != nil {
		t.Errorf("expected nil map for a repo with no .gitmodules, got: %v", urlMap)
	}
}

// TestSubmoduleURLMap_ReturnsLocalPathURL verifies that a submodule tracked
// via a local path has its URL recorded in the map and that the URL is
// correctly identified as a local path by isLocalPath.
func TestSubmoduleURLMap_ReturnsLocalPathURL(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-url-map")

	urlMap, err := submoduleURLMap(umbrella)
	if err != nil {
		t.Fatalf("submoduleURLMap: %v", err)
	}
	if urlMap == nil {
		t.Fatal("expected a non-nil map, got nil")
	}

	url, ok := urlMap[subPath]
	if !ok {
		t.Fatalf("expected map to contain path %q, keys: %v", subPath, urlMap)
	}
	if !isLocalPath(url) {
		t.Errorf("expected URL %q to be a local path, but isLocalPath returned false", url)
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

// TestRunPull_SkipsLocalPathSubmodule verifies that a submodule whose URL is a
// local path is skipped with a hint rather than attempting a remote pull.
func TestRunPull_SkipsLocalPathSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-pull-local")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPull(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPull with local path submodule: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "hint:") {
		t.Errorf("expected stderr to contain 'hint:', got:\n%s", stderr)
	}
	if !strings.Contains(stderr, subPath) {
		t.Errorf("expected stderr to mention submodule %q, got:\n%s", subPath, stderr)
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

// TestRunPull_SkipsLocalPathEvenWithNewCommit verifies that a submodule whose
// URL is a local path is not pulled even when new commits are available in the
// source repository — the hint is printed and HEAD remains unchanged.
func TestRunPull_SkipsLocalPathEvenWithNewCommit(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-pull-local-commit")

	// Rename the default branch to "main" for determinism.
	runGitIn(t, source, "branch", "-M", "main")

	rel := relPath(umbrella, source)
	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "main", tc, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subDir := filepath.Join(umbrella, "svc-pull-local-commit")

	// Check out the named branch so the submodule is not in detached HEAD.
	runGitIn(t, subDir, "checkout", "main")

	before := runGitIn(t, subDir, "rev-parse", "HEAD")

	// Push a new commit to the source repository.
	writeFile(t, filepath.Join(source, "feature.go"), "package main\n// new feature\n")
	runGitIn(t, source, "add", ".")
	runGitIn(t, source, "commit", "-m", "feat: new feature")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPull(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPull: %v", err)
	}

	// HEAD must not have advanced — the local-path submodule was skipped.
	after := runGitIn(t, subDir, "rev-parse", "HEAD")
	if before != after {
		t.Errorf("expected submodule HEAD to remain at %s after skipped pull, but it moved to %s", before, after)
	}

	if !strings.Contains(stderrBuf.String(), "hint:") {
		t.Errorf("expected hint message in stderr, got:\n%s", stderrBuf.String())
	}
}
