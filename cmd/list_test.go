package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
	"github.com/spf13/cobra"
)

func TestRunList_WorkspaceNoRepos(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := t.TempDir()
	initCmd := &cobra.Command{}
	initCmd.SetErr(new(bytes.Buffer))
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var stdout, stderr bytes.Buffer
	listCmd := &cobra.Command{}
	listCmd.SetOut(&stdout)
	listCmd.SetErr(&stderr)
	if err := runList(umbrella, listCmd, nil); err != nil {
		t.Fatalf("runList: %v", err)
	}

	if !strings.Contains(stdout.String(), "no managed repos found") {
		t.Fatalf("expected empty-workspace message, got:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "gyat track") {
		t.Fatalf("expected gyat track hint, got:\n%s", stderr.String())
	}
}

func TestRunList_WorkspaceManifest(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-search-v2")
	initCmd := &cobra.Command{}
	initCmd.SetErr(new(bytes.Buffer))
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", trackCmd, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	var stdout bytes.Buffer
	listCmd := &cobra.Command{}
	listCmd.SetOut(&stdout)
	listCmd.SetErr(new(bytes.Buffer))
	if err := runList(umbrella, listCmd, nil); err != nil {
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

func TestRunList_WithParallelPreservesRowOrder(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepos(t, "svc-list-v2-parallel-a", "svc-list-v2-parallel-b")

	var stdout bytes.Buffer
	listCmd := &cobra.Command{}
	listCmd.SetOut(&stdout)
	listCmd.SetErr(io.Discard)

	if err := runListWithFlags(umbrella, workspaceTargetFlags{parallel: true}, listCmd, nil); err != nil {
		t.Fatalf("runListWithFlags: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 5 {
		t.Fatalf("expected header plus three rows, got:\n%s", stdout.String())
	}
	if !strings.HasPrefix(lines[2], ".") {
		t.Fatalf("expected umbrella row first, got %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "svc-list-v2-parallel-a") {
		t.Fatalf("expected first repo row second, got %q", lines[3])
	}
	if !strings.HasPrefix(lines[4], "svc-list-v2-parallel-b") {
		t.Fatalf("expected second repo row third, got %q", lines[4])
	}
}

func TestRunList_UsesInteractiveTUIWhenEnabled(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-list-v2-ui")

	oldDetector := pagerTerminalDetector
	oldStdin := pagerStdin
	oldRunner := listTUIRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerStdin = oldStdin
		listTUIRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	stdinFile, err := os.CreateTemp(t.TempDir(), "list-stdin")
	if err != nil {
		t.Fatalf("CreateTemp stdin: %v", err)
	}
	defer stdinFile.Close()
	pagerStdin = stdinFile

	stdoutFile, err := os.CreateTemp(t.TempDir(), "list-stdout")
	if err != nil {
		t.Fatalf("CreateTemp stdout: %v", err)
	}
	defer stdoutFile.Close()

	called := false
	listTUIRunner = func(title string, entries []uiModel.RepositoryEntry, in *os.File, out *os.File) error {
		called = true
		if title != "gyat list" {
			t.Fatalf("title = %q, want gyat list", title)
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

	if err := runList(umbrella, cmd, nil); err != nil {
		t.Fatalf("runList: %v", err)
	}
	if !called {
		t.Fatal("expected list TUI runner to be used")
	}
}

func TestRunList_NoUIDisablesInteractiveTUI(t *testing.T) {
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-list-v2-no-ui")

	oldDetector := pagerTerminalDetector
	oldStdin := pagerStdin
	oldRunner := listTUIRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerStdin = oldStdin
		listTUIRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	stdinFile, err := os.CreateTemp(t.TempDir(), "list-stdin")
	if err != nil {
		t.Fatalf("CreateTemp stdin: %v", err)
	}
	defer stdinFile.Close()
	pagerStdin = stdinFile

	called := false
	listTUIRunner = func(string, []uiModel.RepositoryEntry, *os.File, *os.File) error {
		called = true
		return nil
	}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)

	if err := runListWithFlags(umbrella, workspaceTargetFlags{noUI: true}, cmd, nil); err != nil {
		t.Fatalf("runListWithFlags: %v", err)
	}
	if called {
		t.Fatal("expected --no-ui to bypass list TUI runner")
	}
	if !strings.Contains(stdout.String(), "svc-list-v2-no-ui") {
		t.Fatalf("expected plain-text output, got:\n%s", stdout.String())
	}
}
