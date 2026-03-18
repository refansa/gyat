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

// TestRunRemove_HappyPath verifies the full three-step removal: deinit,
// module cache deletion, and git rm.
func TestRunRemove_HappyPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "my-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	subName := filepath.Base(source)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runRemove: %v", err)
	}

	// The submodule working-tree directory must be gone.
	assertPathAbsent(t, filepath.Join(umbrella, subName))
}

// TestRunRemove_CleansGitmodules verifies that the submodule entry is removed
// from .gitmodules after a successful removal.
func TestRunRemove_CleansGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "billing-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	subName := filepath.Base(source)
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), subName)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runRemove: %v", err)
	}

	assertFileNotContains(t, filepath.Join(umbrella, ".gitmodules"), subName)
}

// TestRunRemove_CleansModuleCache verifies that the cached module data stored
// under .git/modules/<path> is deleted during removal.
func TestRunRemove_CleansModuleCache(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "auth-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	subName := filepath.Base(source)
	cacheDir := filepath.Join(umbrella, ".git", "modules", subName)

	// The cache directory is created by git submodule add — confirm it exists
	// before we try to remove it.
	assertPathExists(t, cacheDir)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runRemove: %v", err)
	}

	assertPathAbsent(t, cacheDir)
}

// TestRunRemove_SubdirectoryPath verifies that submodules placed inside a
// subdirectory (e.g. "services/auth") are removed cleanly.
func TestRunRemove_SubdirectoryPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "notifications-service")

	dest := "services/notifications"
	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{relPath(umbrella, source), dest}); err != nil {
		t.Fatalf("setup: runAdd with dest path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, dest))

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{dest}); err != nil {
		t.Fatalf("runRemove with subdirectory path: %v", err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, dest))
}

// TestRunRemove_NonExistentSubmodule verifies that trying to remove a path
// that was never registered as a submodule returns a non-nil error.
func TestRunRemove_NonExistentSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := newUmbrellaRepo(t)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	err := runRemove(umbrella, rc, []string{"does-not-exist"})
	if err == nil {
		t.Fatal("expected an error when removing a non-existent submodule, got nil")
	}
}

// TestRunRemove_PathIsCleaned verifies that paths with redundant separators or
// dot segments are normalized before being passed to git, so that e.g.
// "./my-service" and "my-service" both refer to the same submodule.
func TestRunRemove_PathIsCleaned(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "reporting-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	subName := filepath.Base(source)

	// Pass the path with a leading "./" — filepath.Clean inside runRemove
	// should normalise this to just the base name.
	dirtyPath := "." + string(os.PathSeparator) + subName
	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{dirtyPath}); err != nil {
		t.Fatalf("runRemove with dirty path %q: %v", dirtyPath, err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, subName))
}

// TestRunRemove_RemovesThenAdd verifies that after a submodule is removed, the
// same path can be re-added without error — a common workflow when swapping out
// an implementation repo.
func TestRunRemove_RemovesThenAdd(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "gateway-service")

	rel := relPath(umbrella, source)
	subName := filepath.Base(source)

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: first runAdd: %v", err)
	}

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{subName}); err != nil {
		t.Fatalf("runRemove: %v", err)
	}

	// Re-adding the same source to the same path must succeed.
	ac2 := &cobra.Command{}
	ac2.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac2, []string{rel}); err != nil {
		t.Fatalf("runAdd after remove: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, subName))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), subName)
}

// TestRunRemove_OutputMentionsPath verifies that the success message printed to
// stderr includes the submodule path so the user knows what was removed.
func TestRunRemove_OutputMentionsPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "search-service")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runAdd(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	subName := filepath.Base(source)

	var stderrBuf bytes.Buffer
	rc := &cobra.Command{}
	rc.SetErr(&stderrBuf)
	if err := runRemove(umbrella, rc, []string{subName}); err != nil {
		t.Errorf("runRemove: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), subName) {
		t.Errorf("expected stderr to mention %q, got:\n%s", subName, stderrBuf.String())
	}
}
