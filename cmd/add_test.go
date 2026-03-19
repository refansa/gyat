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
// Helpers specific to staging tests
// ---------------------------------------------------------------------------

// setupTrackedSubmodule creates an umbrella repo with a single tracked
// submodule and returns the umbrella directory and the submodule path within
// it. The source repo is named subName and is placed as a sibling of the
// umbrella directory so that the relative path ../subName is valid.
func setupTrackedSubmodule(t *testing.T, subName string) (umbrella, subPath string) {
	t.Helper()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, subName)
	rel := relPath(umbrella, source)

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{rel}); err != nil {
		t.Fatalf("setupTrackedSubmodule: runTrack %q: %v", rel, err)
	}

	return umbrella, subName
}

// dirtySubmodule writes a new file into subDir, leaving the submodule working
// tree with an untracked file that is ready to be staged.
func dirtySubmodule(t *testing.T, subDir string) {
	t.Helper()
	writeFile(t, filepath.Join(subDir, "new_feature.go"), "package main\n// new feature\n")
}

// stagedFilesInDir returns the newline-separated list of files that are
// currently staged (in the index) inside dir.
func stagedFilesInDir(t *testing.T, dir string) string {
	t.Helper()
	return runGitIn(t, dir, "diff", "--cached", "--name-only")
}

// ---------------------------------------------------------------------------
// Unit tests — allSubmodulePaths
// ---------------------------------------------------------------------------

// TestAllSubmodulePaths_NoGitmodules verifies that a repo with no .gitmodules
// returns an empty slice without error.
func TestAllSubmodulePaths_NoGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	paths, err := allSubmodulePaths(dir)
	if err != nil {
		t.Fatalf("allSubmodulePaths on repo with no .gitmodules: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected no paths, got %v", paths)
	}
}

// TestAllSubmodulePaths_ReturnsPaths verifies that a repo with a tracked
// submodule returns that submodule's path.
func TestAllSubmodulePaths_ReturnsPaths(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-paths")

	paths, err := allSubmodulePaths(umbrella)
	if err != nil {
		t.Fatalf("allSubmodulePaths: %v", err)
	}
	if len(paths) != 1 || paths[0] != subPath {
		t.Errorf("allSubmodulePaths = %v, want [%s]", paths, subPath)
	}
}

// TestAllSubmodulePaths_ReturnsMultiplePaths verifies that all registered
// submodule paths are returned when there is more than one.
func TestAllSubmodulePaths_ReturnsMultiplePaths(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-a")
	svcB := filepath.Join(base, "svc-b")

	for _, d := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, d := range []string{svcA, svcB} {
		runGitIn(t, d, "init")
		runGitIn(t, d, "config", "user.email", "test@gyat.test")
		runGitIn(t, d, "config", "user.name", "gyat test")
		runGitIn(t, d, "config", "commit.gpgsign", "false")
		runGitIn(t, d, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(d, "main.go"), "package main\n")
		runGitIn(t, d, "add", ".")
		runGitIn(t, d, "commit", "-m", "init")
	}

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{"../svc-a"}); err != nil {
		t.Fatalf("track svc-a: %v", err)
	}
	if err := runTrack(umbrella, "", tc, []string{"../svc-b"}); err != nil {
		t.Fatalf("track svc-b: %v", err)
	}

	paths, err := allSubmodulePaths(umbrella)
	if err != nil {
		t.Fatalf("allSubmodulePaths: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %v", paths)
	}

	pathSet := map[string]bool{}
	for _, p := range paths {
		pathSet[p] = true
	}
	for _, want := range []string{"svc-a", "svc-b"} {
		if !pathSet[want] {
			t.Errorf("allSubmodulePaths missing %q, got %v", want, paths)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests — runAdd (staging)
// ---------------------------------------------------------------------------

// TestRunAdd_NoSubmodules_PrintsHint verifies that when there are no
// registered submodules, a friendly hint pointing to 'gyat track' is printed.
func TestRunAdd_NoSubmodules_PrintsHint(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(dir, cmd, nil); err != nil {
		t.Fatalf("runAdd on empty repo: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "gyat track") {
		t.Errorf("expected hint referencing 'gyat track'\ngot: %q", stderrBuf.String())
	}
}

// TestRunAdd_NothingToStage_PrintsMessage verifies that when all submodule
// working trees are clean, a 'nothing to stage' message is printed.
func TestRunAdd_NothingToStage_PrintsMessage(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	// setupTrackedSubmodule leaves the submodule with a clean working tree
	// (only the committed main.go is present).
	umbrella, _ := setupTrackedSubmodule(t, "svc-clean")

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd on clean submodules: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "nothing to stage") {
		t.Errorf("expected 'nothing to stage' message\ngot: %q", stderrBuf.String())
	}
}

// TestRunAdd_StagesAllSubmodules verifies that runAdd with no args stages
// pending changes across all registered submodules.
func TestRunAdd_StagesAllSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-stage-all")
	subDir := filepath.Join(umbrella, subPath)

	dirtySubmodule(t, subDir)

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd: %v", err)
	}

	if staged := stagedFilesInDir(t, subDir); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected 'new_feature.go' to be staged in %s\ngit diff --cached:\n%s", subPath, staged)
	}
}

// TestRunAdd_StagesSpecificSubmodule verifies that when a submodule path is
// explicitly provided, only that submodule is staged and others are left
// untouched.
func TestRunAdd_StagesSpecificSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-a")
	svcB := filepath.Join(base, "svc-b")

	for _, d := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, d := range []string{svcA, svcB} {
		runGitIn(t, d, "init")
		runGitIn(t, d, "config", "user.email", "test@gyat.test")
		runGitIn(t, d, "config", "user.name", "gyat test")
		runGitIn(t, d, "config", "commit.gpgsign", "false")
		runGitIn(t, d, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(d, "main.go"), "package main\n")
		runGitIn(t, d, "add", ".")
		runGitIn(t, d, "commit", "-m", "init")
	}

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{"../svc-a"}); err != nil {
		t.Fatalf("track svc-a: %v", err)
	}
	if err := runTrack(umbrella, "", tc, []string{"../svc-b"}); err != nil {
		t.Fatalf("track svc-b: %v", err)
	}

	// Dirty both submodules.
	dirtySubmodule(t, filepath.Join(umbrella, "svc-a"))
	dirtySubmodule(t, filepath.Join(umbrella, "svc-b"))

	// Stage only svc-a.
	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	if err := runAdd(umbrella, cmd, []string{"svc-a"}); err != nil {
		t.Fatalf("runAdd svc-a: %v", err)
	}

	// svc-a must have staged changes.
	if staged := stagedFilesInDir(t, filepath.Join(umbrella, "svc-a")); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected svc-a to have staged changes\ngot: %q", staged)
	}

	// svc-b must NOT have staged changes.
	if staged := stagedFilesInDir(t, filepath.Join(umbrella, "svc-b")); staged != "" {
		t.Errorf("expected svc-b to have no staged changes, got: %q", staged)
	}
}

// TestRunAdd_UncheckedOutSubmodule_Warns verifies that a submodule whose
// directory is absent on disk (e.g. not yet initialized) is skipped with a
// warning rather than causing a fatal error.
func TestRunAdd_UncheckedOutSubmodule_Warns(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-absent")

	// Remove the checked-out submodule directory to simulate an
	// uninitialized / not-yet-cloned submodule.
	if err := os.RemoveAll(filepath.Join(umbrella, subPath)); err != nil {
		t.Fatalf("remove submodule dir: %v", err)
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd with absent submodule dir should not error: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "warning") {
		t.Errorf("expected a warning about the absent submodule\ngot: %q", stderrBuf.String())
	}
}

// TestRunAdd_MultipleSubmodules_StagesAll verifies that all dirty submodules
// are staged when no path arguments are given.
func TestRunAdd_MultipleSubmodules_StagesAll(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-multi-a")
	svcB := filepath.Join(base, "svc-multi-b")

	for _, d := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, d := range []string{svcA, svcB} {
		runGitIn(t, d, "init")
		runGitIn(t, d, "config", "user.email", "test@gyat.test")
		runGitIn(t, d, "config", "user.name", "gyat test")
		runGitIn(t, d, "config", "commit.gpgsign", "false")
		runGitIn(t, d, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(d, "main.go"), "package main\n")
		runGitIn(t, d, "add", ".")
		runGitIn(t, d, "commit", "-m", "init")
	}

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{"../svc-multi-a"}); err != nil {
		t.Fatalf("track svc-multi-a: %v", err)
	}
	if err := runTrack(umbrella, "", tc, []string{"../svc-multi-b"}); err != nil {
		t.Fatalf("track svc-multi-b: %v", err)
	}

	// Dirty both submodules.
	dirtySubmodule(t, filepath.Join(umbrella, "svc-multi-a"))
	dirtySubmodule(t, filepath.Join(umbrella, "svc-multi-b"))

	// Stage everything with no args.
	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd (no args): %v", err)
	}

	for _, name := range []string{"svc-multi-a", "svc-multi-b"} {
		if staged := stagedFilesInDir(t, filepath.Join(umbrella, name)); !strings.Contains(staged, "new_feature.go") {
			t.Errorf("expected %s to have staged changes\ngot: %q", name, staged)
		}
	}
}
