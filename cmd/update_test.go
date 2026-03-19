package cmd

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunUpdate_NoSubmodules verifies that running update in a repo with no
// submodules exits cleanly without an error. git submodule update is a no-op
// when there is nothing to update.
func TestRunUpdate_NoSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	err := runUpdate(dir, &cobra.Command{}, nil)
	if err != nil {
		t.Fatalf("runUpdate on a repo with no submodules: %v", err)
	}
}

// TestRunUpdate_AllSubmodules verifies that update runs successfully when at
// least one submodule is present and no specific path is given (update all).
func TestRunUpdate_AllSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-update")

	rel := relPath(umbrella, source)
	ac := &cobra.Command{}
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	// The submodule is already at the latest commit, so update is a no-op —
	// but it must still exit cleanly.
	err := runUpdate(umbrella, &cobra.Command{}, nil)
	if err != nil {
		t.Fatalf("runUpdate (all): %v", err)
	}
}

// TestRunUpdate_SpecificPath verifies that passing a submodule path restricts
// the update to only that submodule.
func TestRunUpdate_SpecificPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-specific")

	rel := relPath(umbrella, source)
	ac := &cobra.Command{}
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)
	err := runUpdate(umbrella, &cobra.Command{}, []string{subName})
	if err != nil {
		t.Fatalf("runUpdate (specific path %q): %v", subName, err)
	}
}

// TestRunUpdate_NewCommitIsPickedUp verifies the core purpose of the update
// command: when the source repo has a new commit, runUpdate checks it out.
func TestRunUpdate_NewCommitIsPickedUp(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-ahead")

	rel := relPath(umbrella, source)
	ac := &cobra.Command{}
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	subName := filepath.Base(source)

	// Record the SHA the submodule currently points to.
	before := runGitIn(t, filepath.Join(umbrella, subName), "rev-parse", "HEAD")

	// Push a new commit to the source repo.
	writeFile(t, filepath.Join(source, "newfile.txt"), "hello from new commit\n")
	runGitIn(t, source, "add", ".")
	runGitIn(t, source, "commit", "-m", "second commit")

	// Run update — it should pull the new commit in.
	if err := runUpdate(umbrella, &cobra.Command{}, nil); err != nil {
		t.Fatalf("runUpdate after new commit: %v", err)
	}

	after := runGitIn(t, filepath.Join(umbrella, subName), "rev-parse", "HEAD")

	if before == after {
		t.Errorf(
			"expected submodule HEAD to change after update, but it stayed at %s",
			before,
		)
	}
}

// TestRunUpdate_InvalidPath verifies that passing a path that is not a known
// submodule results in an error.
func TestRunUpdate_InvalidPath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	err := runUpdate(dir, &cobra.Command{}, []string{"does-not-exist"})
	if err == nil {
		t.Error("expected an error when updating a non-existent submodule path, got nil")
	}
}
