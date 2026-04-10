package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/spf13/cobra"
)

func TestRunUntrack_WorkspaceRemovesManifestAndDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "search-service-v2")
	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(io.Discard)
	if err := runTrack(umbrella, "", trackCmd, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)
	if err := runUntrack(umbrella, cmd, []string{"search-service-v2"}); err != nil {
		t.Fatalf("runUntrack: %v", err)
	}

	assertPathAbsent(t, filepath.Join(umbrella, "search-service-v2"))
	assertFileNotContains(t, filepath.Join(umbrella, manifest.FileName), "search-service-v2")
	assertFileNotContains(t, filepath.Join(umbrella, ".gitignore"), "/search-service-v2/")
	if !strings.Contains(stderrBuf.String(), "untracked repository 'search-service-v2'") {
		t.Fatalf("expected v2 untrack message, got:\n%s", stderrBuf.String())
	}
}

func TestRunUntrack_WorkspaceInvalidSelector(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "search-service-v2-invalid")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	err := runUntrack(umbrella, cmd, []string{"ghost-repo"})
	if err == nil {
		t.Fatal("expected an error for an unknown tracked repository")
	}
	if !strings.Contains(err.Error(), "not a tracked repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunUntrack_WithParallelRemovesMultipleRepos(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDirs := setupTrackedWorkspaceRepos(t, "search-service-v2-parallel-a", "search-service-v2-parallel-b")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	flags := workspaceTargetFlags{parallel: true}
	args := []string{"search-service-v2-parallel-a", "search-service-v2-parallel-b"}
	if err := runUntrackWithFlags(umbrella, flags, cmd, args); err != nil {
		t.Fatalf("runUntrackWithFlags: %v", err)
	}

	assertPathAbsent(t, repoDirs["search-service-v2-parallel-a"])
	assertPathAbsent(t, repoDirs["search-service-v2-parallel-b"])
	manifestPath := filepath.Join(umbrella, manifest.FileName)
	assertFileNotContains(t, manifestPath, "search-service-v2-parallel-a")
	assertFileNotContains(t, manifestPath, "search-service-v2-parallel-b")
}
