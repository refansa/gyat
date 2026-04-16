package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
	"github.com/spf13/cobra"
)

func TestParsePorcelain_Empty(t *testing.T) {
	t.Parallel()
	if got := parsePorcelain(""); len(got) != 0 {
		t.Errorf("expected empty slice for empty input, got %v", got)
	}
}

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

func TestParsePorcelain_MultipleEntries(t *testing.T) {
	t.Parallel()
	out := "M  staged.go\n M unstaged.go\n?? untracked.go\n"
	if got := parsePorcelain(out); len(got) != 3 {
		t.Errorf("expected 3 entries, got %d", len(got))
	}
}

func TestParsePorcelain_SkipsShortLines(t *testing.T) {
	t.Parallel()
	out := "M\n??\nM  valid.go\n"
	entries := parsePorcelain(out)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after skipping short lines, got %d: %v", len(entries), entries)
	}
}

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
		t.Run(string(tt.code), func(t *testing.T) {
			t.Parallel()
			if got := statusLabel(tt.code); got != tt.want {
				t.Errorf("statusLabel(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestStatusLabel_UnknownCode(t *testing.T) {
	t.Parallel()
	if got := statusLabel('X'); got != "changed" {
		t.Errorf("statusLabel('X') = %q, want \"changed\"", got)
	}
}

func TestRunStatus_WorkspaceNoRepos(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := t.TempDir()
	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(&stderrBuf)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	if !strings.Contains(stdoutBuf.String(), "umbrella repository") {
		t.Fatalf("expected umbrella repository section, got:\n%s", stdoutBuf.String())
	}
	if !strings.Contains(stderrBuf.String(), "gyat track") {
		t.Fatalf("expected gyat track hint, got:\n%s", stderrBuf.String())
	}
}

func TestRunStatus_WorkspaceAllClean(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-clean")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "umbrella repository") || !strings.Contains(out, "svc-status-v2-clean") {
		t.Fatalf("expected umbrella and repo sections, got:\n%s", out)
	}
	if strings.Count(out, "nothing to commit, working tree clean") != 2 {
		t.Fatalf("expected 2 clean sections, got:\n%s", out)
	}
}

func TestRunStatus_WorkspaceNotCloned(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-status-v2-not-cloned")
	if err := os.RemoveAll(repoDir); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "not cloned") || !strings.Contains(out, "svc-status-v2-not-cloned") {
		t.Fatalf("expected not-cloned section, got:\n%s", out)
	}
}

func TestRunStatus_WorkspaceWithSelectors(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	sourceA := filepath.Join(base, "svc-status-v2-a")
	sourceB := filepath.Join(base, "svc-status-v2-b")
	for _, dir := range []string{umbrella, sourceA, sourceB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, source := range []string{sourceA, sourceB} {
		runGitIn(t, source, "init")
		runGitIn(t, source, "config", "user.email", "test@gyat.test")
		runGitIn(t, source, "config", "user.name", "gyat test")
		runGitIn(t, source, "config", "commit.gpgsign", "false")
		runGitIn(t, source, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(source, "main.go"), "package main\n")
		runGitIn(t, source, "add", ".")
		runGitIn(t, source, "commit", "-m", "initial commit")
	}

	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(io.Discard)
	if err := runTrack(umbrella, "", trackCmd, []string{relPath(umbrella, sourceA)}); err != nil {
		t.Fatalf("runTrack sourceA: %v", err)
	}
	if err := runTrack(umbrella, "", trackCmd, []string{relPath(umbrella, sourceB)}); err != nil {
		t.Fatalf("runTrack sourceB: %v", err)
	}
	commitWorkspaceMetadata(t, umbrella)

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, []string{"svc-status-v2-a"}); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "umbrella repository") || !strings.Contains(out, "svc-status-v2-a") {
		t.Fatalf("expected umbrella and selected repo, got:\n%s", out)
	}
	if strings.Contains(out, "svc-status-v2-b") {
		t.Fatalf("did not expect unselected repo, got:\n%s", out)
	}
}

func TestRunStatus_WithParallelPreservesSectionOrder(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepos(t, "svc-status-v2-parallel-a", "svc-status-v2-parallel-b")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatusWithFlags(umbrella, workspaceTargetFlags{parallel: true}, sc, nil); err != nil {
		t.Fatalf("runStatusWithFlags: %v", err)
	}

	out := stdoutBuf.String()
	rootIndex := strings.Index(out, "umbrella repository")
	repoAIndex := strings.Index(out, "svc-status-v2-parallel-a")
	repoBIndex := strings.Index(out, "svc-status-v2-parallel-b")
	if rootIndex == -1 || repoAIndex == -1 || repoBIndex == -1 {
		t.Fatalf("expected root and repo sections, got:\n%s", out)
	}
	if !(rootIndex < repoAIndex && repoAIndex < repoBIndex) {
		t.Fatalf("expected ordered sections, got:\n%s", out)
	}
	if strings.Count(out, "nothing to commit, working tree clean") != 3 {
		t.Fatalf("expected 3 clean sections, got:\n%s", out)
	}
}

func TestRunStatus_NoPagerFlagPrintsDirectly(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-no-pager")

	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		return "less -FRX", true
	}
	pagerCalled := false
	pagerRunner = func(io.Writer, io.Writer, string, pagerCommand) error {
		pagerCalled = true
		return nil
	}

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)
	sc.Flags().Bool(noPagerFlagName, false, "")
	if err := sc.Flags().Set(noPagerFlagName, "true"); err != nil {
		t.Fatalf("set %s: %v", noPagerFlagName, err)
	}

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if pagerCalled {
		t.Fatal("expected --no-pager to bypass the pager")
	}
	if !strings.Contains(stdoutBuf.String(), "umbrella repository") {
		t.Fatalf("expected direct status output, got:\n%s", stdoutBuf.String())
	}
}

func TestRunStatus_PagesRenderedReport(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-paged")

	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		return "less -FRX", true
	}

	var pagedContent string
	pagerRunner = func(_ io.Writer, _ io.Writer, content string, _ pagerCommand) error {
		pagedContent = content
		return nil
	}

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if pagedContent == "" {
		t.Fatal("expected rendered status report to be sent to pager")
	}
	if !strings.Contains(pagedContent, "umbrella repository") || !strings.Contains(pagedContent, "svc-status-v2-paged") {
		t.Fatalf("unexpected paged content:\n%s", pagedContent)
	}
	if stdoutBuf.Len() != 0 {
		t.Fatalf("expected pager path to bypass direct stdout writes, got:\n%s", stdoutBuf.String())
	}
}

func TestRunStatus_PagedMoreUsesASCIISeparators(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-more")

	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldLookPath := pagerLookPath
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerLookPath = oldLookPath
		pagerRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		return "", false
	}
	pagerLookPath = func(name string) (string, error) {
		if name == "less" {
			return "", errors.New("missing")
		}
		return name, nil
	}

	var pagedContent string
	var usedPager pagerCommand
	pagerRunner = func(_ io.Writer, _ io.Writer, content string, pager pagerCommand) error {
		pagedContent = content
		usedPager = pager
		return nil
	}

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if usedPager.name != "more" {
		t.Fatalf("expected fallback pager 'more', got %#v", usedPager)
	}
	if strings.Contains(pagedContent, "—") || strings.Contains(pagedContent, "─") {
		t.Fatalf("expected ASCII-safe paged content, got:\n%s", pagedContent)
	}
	if !strings.Contains(pagedContent, "umbrella repository - ") {
		t.Fatalf("expected ASCII header separator in paged content, got:\n%s", pagedContent)
	}
}

func TestRunStatus_ChangedOnlyOmitsCleanRepos(t *testing.T) {
	skipIfNoGit(t)

	umbrella, repoDirs := setupTrackedWorkspaceRepos(t, "svc-status-v2-dirty", "svc-status-v2-clean-sibling")
	writeFile(t, filepath.Join(repoDirs["svc-status-v2-dirty"], "dirty.txt"), "dirty\n")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)
	sc.Flags().Bool(changedOnlyFlagName, false, "")
	if err := sc.Flags().Set(changedOnlyFlagName, "true"); err != nil {
		t.Fatalf("set %s: %v", changedOnlyFlagName, err)
	}

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if strings.Contains(out, "umbrella repository") {
		t.Fatalf("did not expect clean umbrella in changed-only output, got:\n%s", out)
	}
	if !strings.Contains(out, "svc-status-v2-dirty") {
		t.Fatalf("expected changed repo in output, got:\n%s", out)
	}
	if strings.Contains(out, "svc-status-v2-clean-sibling") {
		t.Fatalf("did not expect clean repo in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Untracked files:") {
		t.Fatalf("expected dirty repo details, got:\n%s", out)
	}
}

func TestRunStatus_ChangedOnlyReportsWhenAllClean(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-all-clean")

	var stdoutBuf bytes.Buffer
	sc := &cobra.Command{}
	sc.SetOut(&stdoutBuf)
	sc.SetErr(io.Discard)
	sc.Flags().Bool(changedOnlyFlagName, false, "")
	if err := sc.Flags().Set(changedOnlyFlagName, "true"); err != nil {
		t.Fatalf("set %s: %v", changedOnlyFlagName, err)
	}

	if err := runStatus(umbrella, sc, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}

	out := stdoutBuf.String()
	if !strings.Contains(out, "no changed repositories found") {
		t.Fatalf("expected changed-only empty message, got:\n%s", out)
	}
	if strings.Contains(out, "nothing to commit, working tree clean") {
		t.Fatalf("did not expect clean sections in changed-only output, got:\n%s", out)
	}
}

func TestRunStatus_UsesInteractiveTUIWhenEnabled(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-ui")

	oldDetector := pagerTerminalDetector
	oldStdin := pagerStdin
	oldRunner := statusTUIRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerStdin = oldStdin
		statusTUIRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	stdinFile, err := os.CreateTemp(t.TempDir(), "status-stdin")
	if err != nil {
		t.Fatalf("CreateTemp stdin: %v", err)
	}
	defer stdinFile.Close()
	pagerStdin = stdinFile

	stdoutFile, err := os.CreateTemp(t.TempDir(), "status-stdout")
	if err != nil {
		t.Fatalf("CreateTemp stdout: %v", err)
	}
	defer stdoutFile.Close()

	called := false
	statusTUIRunner = func(title string, entries []uiModel.RepositoryEntry, in *os.File, out *os.File) error {
		called = true
		if title != "gyat status" {
			t.Fatalf("title = %q, want gyat status", title)
		}
		if len(entries) != 2 {
			t.Fatalf("entries len = %d, want 2", len(entries))
		}
		if in != stdinFile || out != stdoutFile {
			t.Fatal("unexpected TUI files passed to runner")
		}
		return nil
	}

	cmd := &cobra.Command{}
	cmd.SetOut(stdoutFile)
	cmd.SetErr(io.Discard)

	if err := runStatus(umbrella, cmd, nil); err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if !called {
		t.Fatal("expected status TUI runner to be used")
	}
}

func TestRunStatus_NoUIDisablesInteractiveTUI(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-status-v2-no-ui")

	oldDetector := pagerTerminalDetector
	oldStdin := pagerStdin
	oldRunner := statusTUIRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerStdin = oldStdin
		statusTUIRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	stdinFile, err := os.CreateTemp(t.TempDir(), "status-stdin")
	if err != nil {
		t.Fatalf("CreateTemp stdin: %v", err)
	}
	defer stdinFile.Close()
	pagerStdin = stdinFile

	called := false
	statusTUIRunner = func(string, []uiModel.RepositoryEntry, *os.File, *os.File) error {
		called = true
		return nil
	}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.Flags().Bool(noPagerFlagName, false, "")

	if err := runStatusWithFlags(umbrella, workspaceTargetFlags{noUI: true}, cmd, nil); err != nil {
		t.Fatalf("runStatusWithFlags: %v", err)
	}
	if called {
		t.Fatal("expected --no-ui to bypass status TUI runner")
	}
	if !strings.Contains(stdout.String(), "svc-status-v2-no-ui") {
		t.Fatalf("expected plain-text status output, got:\n%s", stdout.String())
	}
}
