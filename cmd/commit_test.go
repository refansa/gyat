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
// Unit tests — hasStagedChanges
// ---------------------------------------------------------------------------

// TestHasStagedChanges_Empty verifies that an empty status output is
// treated as having no staged changes.
func TestHasStagedChanges_Empty(t *testing.T) {
	t.Parallel()
	if hasStagedChanges("") {
		t.Error("expected false for empty status output")
	}
}

// TestHasStagedChanges_UntrackedOnly verifies that lines starting with '??'
// (untracked files) do not count as staged changes.
func TestHasStagedChanges_UntrackedOnly(t *testing.T) {
	t.Parallel()
	out := "?? newfile.go\n?? another.go\n"
	if hasStagedChanges(out) {
		t.Error("expected false when only untracked files are present")
	}
}

// TestHasStagedChanges_WorkingTreeOnly verifies that lines like ' M file.go'
// (modified in working tree only, not staged) do not count as staged changes.
func TestHasStagedChanges_WorkingTreeOnly(t *testing.T) {
	t.Parallel()
	out := " M file.go\n D removed.go\n"
	if hasStagedChanges(out) {
		t.Error("expected false when changes are only in the working tree (not staged)")
	}
}

// TestHasStagedChanges_StagedFile verifies that lines like 'A  file.go'
// or 'M  file.go' (staged, clean working tree) return true.
func TestHasStagedChanges_StagedFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		out  string
	}{
		{"added", "A  file.go\n"},
		{"modified", "M  file.go\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !hasStagedChanges(tt.out) {
				t.Errorf("expected true for staged file: %q", tt.out)
			}
		})
	}
}

// TestHasStagedChanges_StagedAndModified verifies that lines like 'MM file.go'
// (staged in index AND modified in working tree) return true because X is 'M'.
func TestHasStagedChanges_StagedAndModified(t *testing.T) {
	t.Parallel()
	out := "MM file.go\n"
	if !hasStagedChanges(out) {
		t.Error("expected true when file is staged and also modified in working tree")
	}
}

// TestHasStagedChanges_DeletedInIndex verifies that lines like 'D  file.go'
// (deleted in the index / staged for deletion) return true.
func TestHasStagedChanges_DeletedInIndex(t *testing.T) {
	t.Parallel()
	out := "D  file.go\n"
	if !hasStagedChanges(out) {
		t.Error("expected true when a file is staged for deletion")
	}
}

// ---------------------------------------------------------------------------
// Unit tests — buildCommitArgs
// ---------------------------------------------------------------------------

// TestBuildCommitArgs_Basic verifies that a simple message produces
// ["commit", "-m", "message"].
func TestBuildCommitArgs_Basic(t *testing.T) {
	t.Parallel()
	got := buildCommitArgs("my commit message", false)
	want := []string{"commit", "-m", "my commit message"}
	if len(got) != len(want) {
		t.Fatalf("buildCommitArgs length = %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("buildCommitArgs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestBuildCommitArgs_NoVerify verifies that the --no-verify flag is appended.
func TestBuildCommitArgs_NoVerify(t *testing.T) {
	t.Parallel()
	got := buildCommitArgs("message", true)
	want := []string{"commit", "-m", "message", "--no-verify"}
	if len(got) != len(want) {
		t.Fatalf("buildCommitArgs length = %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("buildCommitArgs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Integration-test helpers
// ---------------------------------------------------------------------------

// commitSetup creates an umbrella repo with a tracked submodule and an initial
// umbrella commit so that subsequent commits can be made. It returns the
// umbrella directory and the absolute path to the submodule working tree.
func commitSetup(t *testing.T, subName string) (umbrella string, subDir string) {
	t.Helper()

	umbrella, _ = setupTrackedSubmodule(t, subName)
	subDir = filepath.Join(umbrella, subName)

	// An initial umbrella commit is required for subsequent commits.
	runGitIn(t, umbrella, "commit", "-m", "track submodule")

	return umbrella, subDir
}

// commitSetupTwo creates an umbrella repo with two tracked submodules and an
// initial umbrella commit. It returns the umbrella dir and both submodule
// working-tree paths.
func commitSetupTwo(t *testing.T, nameA, nameB string) (umbrella, subDirA, subDirB string) {
	t.Helper()
	skipIfNoGit(t)

	// Create umbrella + first source via newTestSetup.
	umbrella, source1 := newTestSetup(t, nameA)
	rel1 := relPath(umbrella, source1)
	tc1 := &cobra.Command{}
	tc1.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc1, []string{rel1}); err != nil {
		t.Fatalf("runTrack %s: %v", nameA, err)
	}

	// Create the second source repo as a sibling of the umbrella.
	base := filepath.Dir(umbrella)
	source2 := filepath.Join(base, nameB)
	if err := os.MkdirAll(source2, 0o755); err != nil {
		t.Fatalf("create %s dir: %v", nameB, err)
	}
	runGitIn(t, source2, "init")
	runGitIn(t, source2, "config", "user.email", "test@gyat.test")
	runGitIn(t, source2, "config", "user.name", "gyat test")
	runGitIn(t, source2, "config", "commit.gpgsign", "false")
	runGitIn(t, source2, "config", "core.autocrlf", "false")
	writeFile(t, filepath.Join(source2, "main.go"), "package main\n")
	runGitIn(t, source2, "add", ".")
	runGitIn(t, source2, "commit", "-m", "initial commit")

	rel2 := relPath(umbrella, source2)
	tc2 := &cobra.Command{}
	tc2.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc2, []string{rel2}); err != nil {
		t.Fatalf("runTrack %s: %v", nameB, err)
	}

	// Initial umbrella commit covering both submodules.
	runGitIn(t, umbrella, "commit", "-m", "track submodules")

	subDirA = filepath.Join(umbrella, nameA)
	subDirB = filepath.Join(umbrella, nameB)

	return umbrella, subDirA, subDirB
}

// ---------------------------------------------------------------------------
// Integration tests — runCommit
// ---------------------------------------------------------------------------

// TestRunCommit_NothingToCommit verifies that when there are no staged changes
// anywhere, runCommit prints "nothing to commit" to stderr.
func TestRunCommit_NothingToCommit(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := commitSetup(t, "svc-commit-nothing")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "should not commit", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "nothing to commit") {
		t.Errorf("expected stderr to contain 'nothing to commit'\ngot: %s", stderrBuf.String())
	}
}

// TestRunCommit_CommitsSubmodule stages a file in a submodule, calls runCommit,
// and verifies the submodule has a new commit with the right message.
func TestRunCommit_CommitsSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := commitSetup(t, "svc-commit-sub")

	// Make and stage a change in the submodule.
	writeFile(t, filepath.Join(subDir, "handler.go"), "package main\n")
	runGitIn(t, subDir, "add", "handler.go")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "feat: add handler", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	// Verify submodule commit message.
	log := runGitIn(t, subDir, "log", "--oneline", "-1")
	if !strings.Contains(log, "feat: add handler") {
		t.Errorf("expected commit message in submodule log\ngot: %s", log)
	}
}

// TestRunCommit_CommitsMultipleSubmodules sets up two submodules, stages
// changes in both, and verifies both get committed.
func TestRunCommit_CommitsMultipleSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDirA, subDirB := commitSetupTwo(t, "svc-multi-a", "svc-multi-b")

	// Stage changes in both submodules.
	writeFile(t, filepath.Join(subDirA, "a.go"), "package a\n")
	runGitIn(t, subDirA, "add", "a.go")

	writeFile(t, filepath.Join(subDirB, "b.go"), "package b\n")
	runGitIn(t, subDirB, "add", "b.go")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "feat: multi commit", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	// Verify both submodules have the commit.
	logA := runGitIn(t, subDirA, "log", "--oneline", "-1")
	if !strings.Contains(logA, "feat: multi commit") {
		t.Errorf("expected commit message in svc-multi-a log\ngot: %s", logA)
	}

	logB := runGitIn(t, subDirB, "log", "--oneline", "-1")
	if !strings.Contains(logB, "feat: multi commit") {
		t.Errorf("expected commit message in svc-multi-b log\ngot: %s", logB)
	}
}

// TestRunCommit_CommitsUmbrellaAfterSubmodules verifies that the umbrella
// repository also gets a commit recording the new submodule refs.
func TestRunCommit_CommitsUmbrellaAfterSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := commitSetup(t, "svc-commit-umbrella")

	// Stage a change in the submodule.
	writeFile(t, filepath.Join(subDir, "service.go"), "package svc\n")
	runGitIn(t, subDir, "add", "service.go")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "chore: update refs", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	// Verify the umbrella also has the commit.
	umbrellaLog := runGitIn(t, umbrella, "log", "--oneline", "-1")
	if !strings.Contains(umbrellaLog, "chore: update refs") {
		t.Errorf("expected commit message in umbrella log\ngot: %s", umbrellaLog)
	}
}

// TestRunCommit_TargetedCommit verifies that when path arguments are passed,
// only the specified submodule gets committed.
func TestRunCommit_TargetedCommit(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDirA, subDirB := commitSetupTwo(t, "svc-targeted-a", "svc-targeted-b")

	// Stage changes in both submodules.
	writeFile(t, filepath.Join(subDirA, "a.go"), "package a\n")
	runGitIn(t, subDirA, "add", "a.go")

	writeFile(t, filepath.Join(subDirB, "b.go"), "package b\n")
	runGitIn(t, subDirB, "add", "b.go")

	// Only commit svc-targeted-a.
	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "feat: targeted", false, cc, []string{"svc-targeted-a"}); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	// svc-targeted-a should have the commit.
	logA := runGitIn(t, subDirA, "log", "--oneline", "-1")
	if !strings.Contains(logA, "feat: targeted") {
		t.Errorf("expected commit message in svc-targeted-a log\ngot: %s", logA)
	}

	// svc-targeted-b should NOT have the commit (its latest should still be
	// "initial commit").
	logB := runGitIn(t, subDirB, "log", "--oneline", "-1")
	if strings.Contains(logB, "feat: targeted") {
		t.Errorf("svc-targeted-b should not have been committed\ngot: %s", logB)
	}
}

// TestRunCommit_SkipsSubmoduleWithNoStagedChanges verifies that a submodule
// with no staged changes is skipped and not committed.
func TestRunCommit_SkipsSubmoduleWithNoStagedChanges(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := commitSetup(t, "svc-commit-skip")

	// Record the HEAD before runCommit to verify it does not change.
	headBefore := runGitIn(t, subDir, "rev-parse", "HEAD")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "should be skipped", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	headAfter := runGitIn(t, subDir, "rev-parse", "HEAD")
	if headBefore != headAfter {
		t.Errorf("submodule HEAD changed despite no staged changes\nbefore: %s\nafter:  %s", headBefore, headAfter)
	}
}

// TestRunCommit_InvalidSubmodulePath verifies that passing a path that is not
// a registered submodule produces an error.
func TestRunCommit_InvalidSubmodulePath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := commitSetup(t, "svc-commit-invalid")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	err := runCommit(umbrella, "bad path", false, cc, []string{"nonexistent-submodule"})
	if err == nil {
		t.Error("expected an error for a non-submodule path, got nil")
	}
}

// TestRunCommit_NoVerifyFlag verifies that the --no-verify flag is correctly
// assembled by buildCommitArgs. We test via buildCommitArgs because actually
// testing hook bypassing in integration is complex and brittle.
func TestRunCommit_NoVerifyFlag(t *testing.T) {
	t.Parallel()

	args := buildCommitArgs("test message", true)

	found := false
	for _, a := range args {
		if a == "--no-verify" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --no-verify in commit args\ngot: %v", args)
	}

	argsWithout := buildCommitArgs("test message", false)
	for _, a := range argsWithout {
		if a == "--no-verify" {
			t.Errorf("did not expect --no-verify when noVerify is false\ngot: %v", argsWithout)
			break
		}
	}
}
