package cmd

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunSync_EmptyRepo verifies that syncing in a repo with no submodules
// exits cleanly — both git submodule sync and git submodule update are
// no-ops when there is nothing to act on.
func TestRunSync_EmptyRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)
	sc := &cobra.Command{}
	sc.SetErr(io.Discard)

	if err := runSync(dir, sc, nil); err != nil {
		t.Fatalf("runSync in empty repo: %v", err)
	}
}

// TestRunSync_WithSubmodule verifies that sync succeeds when at least one
// submodule is registered, leaving the submodule directory intact.
func TestRunSync_WithSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-sync")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	rel := relPath(umbrella, source)
	if err := runAdd(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	sc := &cobra.Command{}
	sc.SetErr(io.Discard)
	if err := runSync(umbrella, sc, nil); err != nil {
		t.Fatalf("runSync with submodule: %v", err)
	}

	// The submodule directory must still be present after a sync.
	assertPathExists(t, filepath.Join(umbrella, filepath.Base(source)))
}

// TestRunSync_OutsideGitRepo verifies that running sync outside any git
// repository returns an error rather than silently succeeding.
func TestRunSync_OutsideGitRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	sc := &cobra.Command{}
	sc.SetErr(io.Discard)

	if err := runSync(dir, sc, nil); err == nil {
		t.Fatal("expected error when running sync outside a git repo, got nil")
	}
}

// TestRunSync_GitmodulesURLIsPreserved verifies that after a sync the URL
// recorded in .gitmodules has not been corrupted or removed.
func TestRunSync_GitmodulesURLIsPreserved(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-url-check")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	rel := relPath(umbrella, source)
	if err := runAdd(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: runAdd: %v", err)
	}

	sc := &cobra.Command{}
	sc.SetErr(io.Discard)
	if err := runSync(umbrella, sc, nil); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	// .gitmodules must still contain a URL entry for the submodule.
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "url")
}

// TestRunSync_MultipleSubmodules verifies that sync handles more than one
// submodule without error.
func TestRunSync_MultipleSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, srcA := newTestSetup(t, "svc-alpha")
	srcB := newSourceRepo(t)

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)

	relA := relPath(umbrella, srcA)
	if err := runAdd(umbrella, "", ac, []string{relA, "services/alpha"}); err != nil {
		t.Fatalf("setup: add first submodule: %v", err)
	}

	relB := relPath(umbrella, srcB)
	if err := runAdd(umbrella, "", ac, []string{relB, "services/beta"}); err != nil {
		t.Fatalf("setup: add second submodule: %v", err)
	}

	sc := &cobra.Command{}
	sc.SetErr(io.Discard)
	if err := runSync(umbrella, sc, nil); err != nil {
		t.Fatalf("runSync with multiple submodules: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, "services/alpha"))
	assertPathExists(t, filepath.Join(umbrella, "services/beta"))
}
