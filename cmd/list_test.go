package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/internal/manifest"
	"github.com/spf13/cobra"
)

// TestRunList_NoGitmodules verifies that list exits cleanly and prints a
// helpful message when no .gitmodules file is present.
func TestRunList_NoGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(dir, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "no submodules found") {
		t.Errorf("expected 'no submodules found' message, got:\n%s", stdout.String())
	}
}

// TestRunList_WithSubmodule verifies that a submodule added via runTrack appears
// in the table output with its path, URL, and a status column.
func TestRunList_WithSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth")

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "service-auth") {
		t.Errorf("expected submodule path 'service-auth' in output, got:\n%s", stdout.String())
	}
}

// TestRunList_HeaderIsPresent verifies that the table header row is always
// printed when at least one submodule is registered.
func TestRunList_HeaderIsPresent(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-billing")

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	for _, col := range []string{"PATH", "BRANCH", "COMMIT", "STATUS", "URL"} {
		if !strings.Contains(stdout.String(), col) {
			t.Errorf("expected column header %q in output, got:\n%s", col, stdout.String())
		}
	}
}

// TestRunList_ShowsURL verifies that the submodule URL from .gitmodules appears
// in the list output.
func TestRunList_ShowsURL(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-payments")

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), filepath.Base(source)) {
		t.Errorf(
			"expected source directory name %q somewhere in URL column, got:\n%s",
			filepath.Base(source), stdout.String(),
		)
	}
}

// TestRunList_DefaultBranchLabel verifies that submodules without an explicit
// tracked branch show the "(default)" label rather than an empty cell.
func TestRunList_DefaultBranchLabel(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-notifications")

	// Track without --branch so no branch is recorded in .gitmodules.
	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "(default)") {
		t.Errorf("expected '(default)' branch label for untracked submodule, got:\n%s", stdout.String())
	}
}

// TestRunList_TrackedBranchShown verifies that a submodule added with --branch
// displays that branch name in the output.
func TestRunList_TrackedBranchShown(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-gateway")

	branch := runGitIn(t, source, "rev-parse", "--abbrev-ref", "HEAD")

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, branch, ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack --branch %s: %v", branch, err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), branch) {
		t.Errorf("expected branch name %q in list output, got:\n%s", branch, stdout.String())
	}
}

// TestRunList_MultipleSubmodules verifies that all added submodules appear in
// the output, each on its own line.
func TestRunList_MultipleSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source1 := newTestSetup(t, "service-users")
	source2 := newSourceRepo(t)

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source1), "service-users"}); err != nil {
		t.Fatalf("setup: runTrack service-users: %v", err)
	}
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source2), "service-orders"}); err != nil {
		t.Fatalf("setup: runTrack service-orders: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	for _, name := range []string{"service-users", "service-orders"} {
		if !strings.Contains(stdout.String(), name) {
			t.Errorf("expected %q in list output, got:\n%s", name, stdout.String())
		}
	}
}

// TestRunList_StatusColumn verifies that a status value is present for each
// listed submodule (any non-empty value in the STATUS column is acceptable).
func TestRunList_StatusColumn(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-search")

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setup: runTrack: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))

	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("runList returned unexpected error: %v", err)
	}

	knownStatuses := []string{"up to date", "not initialized", "modified", "merge conflict", "unknown"}
	found := false
	for _, s := range knownStatuses {
		if strings.Contains(stdout.String(), s) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a known status value in list output, got:\n%s", stdout.String())
	}
}

func TestRunList_WorkspaceManifest(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-search-v2")
	ic := &cobra.Command{}
	ic.SetErr(new(bytes.Buffer))
	if err := runInit(umbrella, ic, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	ac := &cobra.Command{}
	ac.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", ac, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	var stdout bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&stdout)
	lc.SetErr(new(bytes.Buffer))
	if err := runList(umbrella, lc, nil); err != nil {
		t.Fatalf("runList: %v", err)
	}

	if !strings.Contains(stdout.String(), "service-search-v2") {
		t.Fatalf("expected tracked repo path in output, got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "clean") {
		t.Fatalf("expected clean status in output, got:\n%s", stdout.String())
	}
	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "service-search-v2")
}
