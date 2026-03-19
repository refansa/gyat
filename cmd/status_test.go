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
// Unit tests — parsePorcelain
// ---------------------------------------------------------------------------

// TestParsePorcelain_Empty verifies that an empty string produces no entries.
func TestParsePorcelain_Empty(t *testing.T) {
	t.Parallel()
	if got := parsePorcelain(""); len(got) != 0 {
		t.Errorf("expected empty slice for empty input, got %v", got)
	}
}

// TestParsePorcelain_StagedModified verifies that "M  file.go" is parsed as a
// staged modification (X='M', Y=' ').
func TestParsePorcelain_StagedModified(t *testing.T) {
	t.Parallel()
	entries := parsePorcelain("M  handler.go\n")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.x != 'M' || e.y != ' ' || e.path != "handler.go" {
		t.Errorf("unexpected entry: x=%q y=%q path=%q", e.x, e.y, e.path)
	}
}

// TestParsePorcelain_WorkingTreeModified verifies that " M file.go" is parsed
// as a working-tree-only modification (X=' ', Y='M').
func TestParsePorcelain_WorkingTreeModified(t *testing.T) {
	t.Parallel()
	entries := parsePorcelain(" M main.go\n")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.x != ' ' || e.y != 'M' || e.path != "main.go" {
		t.Errorf("unexpected entry: x=%q y=%q path=%q", e.x, e.y, e.path)
	}
}

// TestParsePorcelain_Untracked verifies that "?? file.go" is parsed as an
// untracked entry (X='?', Y='?').
func TestParsePorcelain_Untracked(t *testing.T) {
	t.Parallel()
	entries := parsePorcelain("?? new-file.go\n")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.x != '?' || e.y != '?' || e.path != "new-file.go" {
		t.Errorf("unexpected entry: x=%q y=%q path=%q", e.x, e.y, e.path)
	}
}

// TestParsePorcelain_MultipleEntries verifies that a multi-line output is
// split into the correct number of entries.
func TestParsePorcelain_MultipleEntries(t *testing.T) {
	t.Parallel()
	out := "M  staged.go\n M unstaged.go\n?? untracked.go\n"
	if got := parsePorcelain(out); len(got) != 3 {
		t.Errorf("expected 3 entries, got %d", len(got))
	}
}

// TestParsePorcelain_SkipsShortLines verifies that lines shorter than four
// bytes are silently ignored.
func TestParsePorcelain_SkipsShortLines(t *testing.T) {
	t.Parallel()
	out := "M\n??\nM  valid.go\n"
	entries := parsePorcelain(out)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after skipping short lines, got %d: %v", len(entries), entries)
	}
}

// TestParsePorcelain_CRLFLineEndings verifies that Windows-style CRLF line
// endings are handled and do not bleed into the path field.
func TestParsePorcelain_CRLFLineEndings(t *testing.T) {
	t.Parallel()
	entries := parsePorcelain("M  file.go\r\n?? other.go\r\n")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with CRLF line endings, got %d", len(entries))
	}
	if entries[0].path != "file.go" {
		t.Errorf("expected path 'file.go', got %q", entries[0].path)
	}
	if entries[1].path != "other.go" {
		t.Errorf("expected path 'other.go', got %q", entries[1].path)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — statusLabel
// ---------------------------------------------------------------------------

// TestStatusLabel_KnownCodes verifies that every recognised porcelain status
// code maps to the expected human-readable label.
func TestStatusLabel_KnownCodes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code byte
		want string
	}{
		{'A', "new file"},
		{'M', "modified"},
		{'D', "deleted"},
		{'R', "renamed"},
		{'C', "copied"},
		{'U', "conflict"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.code), func(t *testing.T) {
			t.Parallel()
			if got := statusLabel(tt.code); got != tt.want {
				t.Errorf("statusLabel(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

// TestStatusLabel_UnknownCode verifies that an unrecognised code falls back
// to "changed".
func TestStatusLabel_UnknownCode(t *testing.T) {
	t.Parallel()
	if got := statusLabel('X'); got != "changed" {
		t.Errorf("statusLabel('X') = %q, want \"changed\"", got)
	}
}

// ---------------------------------------------------------------------------
// Integration-test helpers
// ---------------------------------------------------------------------------

// statusSetup creates an umbrella with one tracked submodule and an initial
// umbrella commit, leaving all repositories in a clean state. It returns the
// umbrella directory and the absolute path to the submodule working tree.
func statusSetup(t *testing.T, subName string) (umbrella, subDir string) {
	t.Helper()

	umbrella, _ = setupTrackedSubmodule(t, subName)
	subDir = filepath.Join(umbrella, subName)

	// Commit the umbrella so it starts in a clean, committed state.
	runGitIn(t, umbrella, "commit", "-m", "track submodule")

	return umbrella, subDir
}

// ---------------------------------------------------------------------------
// Integration tests — runStatus
// ---------------------------------------------------------------------------

// TestRunStatus_NoSubmodules verifies that when no submodules have been
// tracked the umbrella section is printed and a hint is written to stderr.
func TestRunStatus_NoSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := newUmbrellaRepo(t)

	var stdoutBuf, stderrBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(&stderrBuf)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	if !strings.Contains(stdoutBuf.String(), "umbrella repository") {
		t.Errorf("expected 'umbrella repository' in stdout\ngot:\n%s", stdoutBuf.String())
	}
	if !strings.Contains(stderrBuf.String(), "hint:") {
		t.Errorf("expected hint in stderr\ngot:\n%s", stderrBuf.String())
	}
	if !strings.Contains(stderrBuf.String(), "gyat track") {
		t.Errorf("expected 'gyat track' hint in stderr\ngot:\n%s", stderrBuf.String())
	}
}

// TestRunStatus_AllClean verifies that when all repositories are clean each
// section reports "nothing to commit, working tree clean".
func TestRunStatus_AllClean(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := statusSetup(t, "svc-status-clean")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "umbrella repository") {
		t.Errorf("expected umbrella repository section\ngot:\n%s", out)
	}
	if !strings.Contains(out, "svc-status-clean") {
		t.Errorf("expected submodule section 'svc-status-clean'\ngot:\n%s", out)
	}
	if strings.Count(out, "nothing to commit, working tree clean") != 2 {
		t.Errorf("expected 2 'nothing to commit' messages (one per repo)\ngot:\n%s", out)
	}
}

// TestRunStatus_StagedChangesInSubmodule verifies that a file staged inside a
// submodule appears under "Changes to be committed" in that submodule's section.
func TestRunStatus_StagedChangesInSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := statusSetup(t, "svc-status-staged")

	// Add and stage a new file inside the submodule.
	writeFile(t, filepath.Join(subDir, "handler.go"), "package main\n")
	runGitIn(t, subDir, "add", "handler.go")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "Changes to be committed:") {
		t.Errorf("expected 'Changes to be committed:' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "handler.go") {
		t.Errorf("expected 'handler.go' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "new file:") {
		t.Errorf("expected 'new file:' label in output\ngot:\n%s", out)
	}
}

// TestRunStatus_UnstagedChangesInSubmodule verifies that an unstaged
// modification inside a submodule appears under "Changes not staged for
// commit" in that submodule's section.
func TestRunStatus_UnstagedChangesInSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := statusSetup(t, "svc-status-unstaged")

	// Overwrite an existing tracked file without staging the change.
	writeFile(t, filepath.Join(subDir, "main.go"), "package main\n// modified\n")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "Changes not staged for commit:") {
		t.Errorf("expected 'Changes not staged for commit:' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "main.go") {
		t.Errorf("expected 'main.go' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "modified:") {
		t.Errorf("expected 'modified:' label in output\ngot:\n%s", out)
	}
}

// TestRunStatus_UntrackedFilesInSubmodule verifies that an untracked file
// inside a submodule appears under "Untracked files" in that section.
func TestRunStatus_UntrackedFilesInSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := statusSetup(t, "svc-status-untracked")

	// Drop a new file into the submodule without staging it.
	writeFile(t, filepath.Join(subDir, "service.go"), "package main\n")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "Untracked files:") {
		t.Errorf("expected 'Untracked files:' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "service.go") {
		t.Errorf("expected 'service.go' in output\ngot:\n%s", out)
	}
}

// TestRunStatus_UmbrellaHasStagedChanges verifies that staged changes in the
// umbrella repository (e.g. right after gyat track, before the first commit)
// are reported in the umbrella section.
func TestRunStatus_UmbrellaHasStagedChanges(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	// setupTrackedSubmodule does NOT commit the umbrella, so .gitmodules and
	// the submodule directory entry are staged but uncommitted.
	umbrella, _ := setupTrackedSubmodule(t, "svc-status-umb-staged")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	// The umbrella section must report staged changes.
	if !strings.Contains(out, "Changes to be committed:") {
		t.Errorf("expected 'Changes to be committed:' in umbrella section\ngot:\n%s", out)
	}
	// .gitmodules is always staged after gyat track.
	if !strings.Contains(out, ".gitmodules") {
		t.Errorf("expected '.gitmodules' among staged changes\ngot:\n%s", out)
	}
}

// TestRunStatus_NotInitialized verifies that a submodule registered in
// .gitmodules whose working-tree directory does not exist is reported as
// "not initialized".
func TestRunStatus_NotInitialized(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := statusSetup(t, "svc-status-noinit")

	// Remove the submodule directory to simulate a not-yet-initialized state.
	if err := os.RemoveAll(subDir); err != nil {
		t.Fatalf("removing submodule dir: %v", err)
	}

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "not initialized") {
		t.Errorf("expected 'not initialized' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "svc-status-noinit") {
		t.Errorf("expected submodule name in output\ngot:\n%s", out)
	}
}

// TestRunStatus_WithPathArgs verifies that passing submodule path arguments
// limits the output to those submodules; the umbrella is always shown first.
func TestRunStatus_WithPathArgs(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _, _ := commitSetupTwo(t, "svc-status-filter-a", "svc-status-filter-b")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	// Request status only for svc-status-filter-a.
	if err := runStatus(umbrella, sc, []string{"svc-status-filter-a"}); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "umbrella repository") {
		t.Errorf("expected umbrella section even when path args are given\ngot:\n%s", out)
	}
	if !strings.Contains(out, "svc-status-filter-a") {
		t.Errorf("expected 'svc-status-filter-a' section\ngot:\n%s", out)
	}
	if strings.Contains(out, "svc-status-filter-b") {
		t.Errorf("did not expect 'svc-status-filter-b' when not requested\ngot:\n%s", out)
	}
}

// TestRunStatus_InvalidPathArg verifies that a path argument that is not a
// registered submodule returns an error.
func TestRunStatus_InvalidPathArg(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := statusSetup(t, "svc-status-invalid")

	sc := &cobra.Command{}
	sc.SetOut(io.Discard)
	sc.SetErr(io.Discard)

	err := runStatus(umbrella, sc, []string{"ghost-service"})
	if err == nil {
		t.Error("expected an error for an unregistered submodule path, got nil")
	}
}

// TestRunStatus_SectionHeaderContainsBranch verifies that the section header
// for each repository includes the current branch name.
func TestRunStatus_SectionHeaderContainsBranch(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := statusSetup(t, "svc-status-branch")

	// Rename the umbrella's branch to something recognisable.
	runGitIn(t, umbrella, "branch", "-m", "feat/my-feature")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	if !strings.Contains(stdoutBuf.String(), "feat/my-feature") {
		t.Errorf("expected branch name 'feat/my-feature' in output\ngot:\n%s", stdoutBuf.String())
	}
}

// TestRunStatus_StagedAndUnstaged verifies that staged index changes and
// unstaged working-tree changes in the same submodule appear in their
// respective sections.
func TestRunStatus_StagedAndUnstaged(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subDir := statusSetup(t, "svc-status-mixed")

	// Stage a new file.
	writeFile(t, filepath.Join(subDir, "api.go"), "package main\n")
	runGitIn(t, subDir, "add", "api.go")

	// Modify an existing tracked file without staging it.
	writeFile(t, filepath.Join(subDir, "main.go"), "package main\n// changed\n")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "Changes to be committed:") {
		t.Errorf("expected 'Changes to be committed:' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "Changes not staged for commit:") {
		t.Errorf("expected 'Changes not staged for commit:' in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "api.go") {
		t.Errorf("expected 'api.go' (staged) in output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "main.go") {
		t.Errorf("expected 'main.go' (unstaged) in output\ngot:\n%s", out)
	}
}
