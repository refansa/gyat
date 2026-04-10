package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/internal/manifest"
	"github.com/spf13/cobra"
)

// TestRunUntrack_HappyPath verifies the full three-step untrack: deinit,
// module cache deletion, and git rm.
func TestRunUntrack_HappyPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "my-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runUntrack(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runUntrack: %v", err)
	}

	// The submodule working-tree directory must be gone.
	assertPathAbsent(t, filepath.Join(umbrella, subName))
}

// TestRunUntrack_CleansGitmodules verifies that the submodule entry is removed
// from .gitmodules after a successful untrack.
func TestRunUntrack_CleansGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "billing-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), subName)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runUntrack(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runUntrack: %v", err)
	}

	assertFileNotContains(t, filepath.Join(umbrella, ".gitmodules"), subName)
}

// TestRunUntrack_CleansModuleCache verifies that the cached module data stored
// under .git/modules/<path> is deleted during untrack.
func TestRunUntrack_CleansModuleCache(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "auth-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)
	cacheDir := filepath.Join(umbrella, ".git", "modules", subName)

	// The cache directory is created by git submodule add — confirm it exists
	// before we try to remove it.
	assertPathExists(t, cacheDir)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runUntrack(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runUntrack: %v", err)
	}

	assertPathAbsent(t, cacheDir)
}

// TestRunUntrack_SubdirectoryPath verifies that submodules placed inside a
// subdirectory (e.g. "services/auth") are untracked cleanly.
func TestRunUntrack_SubdirectoryPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "notifications-service")

	dest := "services/notifications"
	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source), dest}); err != nil {
		t.Fatalf("setup: runTrack with dest path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, dest))

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runUntrack(umbrella, rc, []string{dest}); err != nil {
		t.Fatalf("runUntrack with subdirectory path: %v", err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, dest))
}

// TestRunUntrack_NonExistentSubmodule verifies that trying to untrack a path
// that was never registered as a submodule returns a non-nil error.
func TestRunUntrack_NonExistentSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := newUmbrellaRepo(t)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	err := runUntrack(umbrella, rc, []string{"does-not-exist"})
	if err == nil {
		t.Fatal("expected an error when untracking a non-existent submodule, got nil")
	}
}

// TestRunUntrack_PathIsCleaned verifies that paths with redundant separators or
// dot segments are normalized before being passed to git, so that e.g.
// "./my-service" and "my-service" both refer to the same submodule.
func TestRunUntrack_PathIsCleaned(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "reporting-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)

	// Pass the path with a leading "./" — filepath.Clean inside runUntrack
	// should normalise this to just the base name.
	dirtyPath := "." + string(os.PathSeparator) + subName
	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runUntrack(umbrella, rc, []string{dirtyPath}); err != nil {
		t.Fatalf("runUntrack with dirty path %q: %v", dirtyPath, err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, subName))
}

// TestRunUntrack_RemovesThenTrack verifies that after a submodule is untracked, the
// same path can be re-tracked without error — a common workflow when swapping out
// an implementation repo.
func TestRunUntrack_RemovesThenTrack(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "gateway-service")

	rel := relPath(umbrella, source)
	subName := filepath.Base(source)

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: first runTrack: %v", err)
	}

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runUntrack(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runUntrack: %v", err)
	}

	// Re-tracking the same source to the same path must succeed.
	ac2 := &cobra.Command{}
	ac2.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac2, []string{rel}); err != nil {
		t.Fatalf("runTrack after untrack: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, subName))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), subName)
}

// TestRunUntrack_OutputMentionsPath verifies that the success message printed to
// stderr includes the submodule path so the user knows what was untracked.
func TestRunUntrack_OutputMentionsPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "search-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)

	var stderrBuf bytes.Buffer
	rc := &cobra.Command{}
	rc.SetErr(&stderrBuf)
	if err := runUntrack(umbrella, rc, []string{subName}); err != nil {
		t.Errorf("runUntrack: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), subName) {
		t.Errorf("expected stderr to mention %q, got:\n%s", subName, stderrBuf.String())
	}
}

func TestRunUntrack_WorkspaceRemovesManifestAndDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "search-service-v2")
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)
	if err := runInit(umbrella, ic, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	var stderrBuf bytes.Buffer
	rc := &cobra.Command{}
	rc.SetErr(&stderrBuf)
	if err := runUntrack(umbrella, rc, []string{"search-service-v2"}); err != nil {
		t.Fatalf("runUntrack: %v", err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, "search-service-v2"))
	assertFileNotContains(t, filepath.Join(umbrella, manifest.FileName), "search-service-v2")
	assertFileNotContains(t, filepath.Join(umbrella, ".gitignore"), "/search-service-v2/")
	if !strings.Contains(stderrBuf.String(), "untracked repository 'search-service-v2'") {
		t.Fatalf("expected v2 untrack message, got:\n%s", stderrBuf.String())
	}
}
