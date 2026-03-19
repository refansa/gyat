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

// TestIntegration_FullWorkflow exercises the entire gyat lifecycle in order:
// init → track → list → remove. It is the closest thing to a real user session
// without touching the network.
func TestIntegration_FullWorkflow(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "my-service")

	// ── 1. gyat init ──────────────────────────────────────────────────────────

	ic := &cobra.Command{}
	ic.SetErr(io.Discard)
	if err := runInit(umbrella, ic, nil); err != nil {
		t.Fatalf("step 1 (gyat init): %v", err)
	}
	assertPathExists(t, filepath.Join(umbrella, ".git"))

	// ── 2. gyat track ─────────────────────────────────────────────────────────

	rel := relPath(umbrella, source)
	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("step 2 (gyat track %s): %v", rel, err)
	}
	assertPathExists(t, filepath.Join(umbrella, ".gitmodules"))
	assertPathExists(t, filepath.Join(umbrella, "my-service"))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "my-service")

	// ── 3. gyat list ──────────────────────────────────────────────────────────

	var listOut bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&listOut)
	lc.SetErr(new(bytes.Buffer))
	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("step 3 (gyat list): %v", err)
	}

	out := listOut.String()
	if !strings.Contains(out, "my-service") {
		t.Errorf("step 3: expected 'my-service' in list output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "PATH") {
		t.Errorf("step 3: expected table header in list output\ngot:\n%s", out)
	}

	// ── 4. gyat remove ────────────────────────────────────────────────────────

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{"my-service"}); err != nil {
		t.Fatalf("step 4 (gyat remove my-service): %v", err)
	}
	assertPathAbsent(t, filepath.Join(umbrella, "my-service"))
	assertPathAbsent(t, filepath.Join(umbrella, ".git", "modules", "my-service"))
	assertFileNotContains(t, filepath.Join(umbrella, ".gitmodules"), "my-service")
}

// TestIntegration_TrackWithExplicitDestination verifies that gyat track correctly
// places the submodule under the requested destination path instead of the
// default one derived from the repo name.
func TestIntegration_TrackWithExplicitDestination(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "auth")

	rel := relPath(umbrella, source)
	dest := "services/auth"

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{rel, dest}); err != nil {
		t.Fatalf("gyat track with destination: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, "services", "auth"))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "services/auth")
}

// TestIntegration_TrackWithBranch verifies that the --branch flag is recorded in
// .gitmodules so that gyat update later tracks the right branch.
func TestIntegration_TrackWithBranch(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "billing")

	// Rename the default branch to "main" in the source repo so the flag has
	// something real to point at.
	runGitIn(t, source, "branch", "-m", "main")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	rel := relPath(umbrella, source)
	if err := runTrack(umbrella, "main", ac, []string{rel}); err != nil {
		t.Fatalf("gyat track --branch main: %v", err)
	}

	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "branch = main")
}

// TestIntegration_TrackMultipleSubmodules verifies that gyat list shows all
// registered submodules when more than one has been tracked.
func TestIntegration_TrackMultipleSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()

	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-a")
	svcB := filepath.Join(base, "svc-b")

	for _, dir := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	// Umbrella
	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	// Service A
	runGitIn(t, svcA, "init")
	runGitIn(t, svcA, "config", "user.email", "test@gyat.test")
	runGitIn(t, svcA, "config", "user.name", "gyat test")
	runGitIn(t, svcA, "config", "commit.gpgsign", "false")
	runGitIn(t, svcA, "config", "core.autocrlf", "false")
	writeFile(t, filepath.Join(svcA, "go.mod"), "module svc-a\n")
	runGitIn(t, svcA, "add", ".")
	runGitIn(t, svcA, "commit", "-m", "init svc-a")

	// Service B
	runGitIn(t, svcB, "init")
	runGitIn(t, svcB, "config", "user.email", "test@gyat.test")
	runGitIn(t, svcB, "config", "user.name", "gyat test")
	runGitIn(t, svcB, "config", "commit.gpgsign", "false")
	runGitIn(t, svcB, "config", "core.autocrlf", "false")
	writeFile(t, filepath.Join(svcB, "go.mod"), "module svc-b\n")
	runGitIn(t, svcB, "add", ".")
	runGitIn(t, svcB, "commit", "-m", "init svc-b")

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)
	if err := runTrack(umbrella, "", ac, []string{"../svc-a"}); err != nil {
		t.Fatalf("track svc-a: %v", err)
	}
	if err := runTrack(umbrella, "", ac, []string{"../svc-b"}); err != nil {
		t.Fatalf("track svc-b: %v", err)
	}

	var listOut bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&listOut)
	lc.SetErr(new(bytes.Buffer))
	if err := runList(umbrella, lc, nil); err != nil {
		t.Errorf("gyat list: %v", err)
	}

	out := listOut.String()
	for _, name := range []string{"svc-a", "svc-b"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output\ngot:\n%s", name, out)
		}
	}
}

// TestIntegration_RemoveThenReTrack verifies that a submodule can be removed and
// then re-tracked cleanly with no leftover state from the first addition.
func TestIntegration_RemoveThenReTrack(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "notifications")

	rel := relPath(umbrella, source)

	ac := &cobra.Command{}
	ac.SetErr(io.Discard)

	// First track.
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("first track: %v", err)
	}

	// Remove it.
	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	if err := runRemove(umbrella, rc, []string{"notifications"}); err != nil {
		t.Fatalf("remove: %v", err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, "notifications"))
	assertPathAbsent(t, filepath.Join(umbrella, ".git", "modules", "notifications"))

	// Re-track it — must succeed without leftover state causing a conflict.
	if err := runTrack(umbrella, "", ac, []string{rel}); err != nil {
		t.Fatalf("re-track after remove: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, "notifications"))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "notifications")
}

// TestIntegration_ListWithNoSubmodules verifies that gyat list exits cleanly
// and prints a friendly message when no submodules have been added yet.
func TestIntegration_ListWithNoSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	var listOut bytes.Buffer
	lc := &cobra.Command{}
	lc.SetOut(&listOut)
	lc.SetErr(new(bytes.Buffer))
	if err := runList(dir, lc, nil); err != nil {
		t.Errorf("gyat list on empty repo: %v", err)
	}

	if !strings.Contains(listOut.String(), "no submodules found") {
		t.Errorf("expected 'no submodules found' message\ngot:\n%s", listOut.String())
	}
}

// TestIntegration_RemoveNonExistentSubmodule verifies that removing a path
// that was never added as a submodule returns an error rather than silently
// succeeding.
func TestIntegration_RemoveNonExistentSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	rc := &cobra.Command{}
	rc.SetErr(io.Discard)
	err := runRemove(dir, rc, []string{"ghost-service"})
	if err == nil {
		t.Error("expected error when removing a non-existent submodule, got nil")
	}
}

// TestIntegration_AbsolutePathWarning verifies that tracking a submodule via an
// absolute local path still works but prints a portability warning to stderr.
func TestIntegration_AbsolutePathWarning(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "payments")

	var stderrBuf bytes.Buffer
	ac := &cobra.Command{}
	ac.SetErr(&stderrBuf)

	// source is already an absolute path.
	// Ignore the error: on some git versions the add may also fail with a
	// security error for absolute paths; we are only testing the warning.
	_ = runTrack(umbrella, "", ac, []string{source})

	if !strings.Contains(stderrBuf.String(), "absolute path") {
		t.Errorf("expected portability warning about absolute path in stderr\ngot: %q", stderrBuf.String())
	}
}
