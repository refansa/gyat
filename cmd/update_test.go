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

func TestRunUpdate_WorkspacePullsManifestBranch(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-update-v2-branch")
	runGitIn(t, source, "branch", "-m", "main")

	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(io.Discard)
	if err := runTrack(umbrella, "main", trackCmd, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}
	commitWorkspaceMetadata(t, umbrella)

	repoDir := filepath.Join(umbrella, "svc-update-v2-branch")
	headBefore := runGitIn(t, repoDir, "rev-parse", "HEAD")

	writeFile(t, filepath.Join(source, "new_feature.go"), "package main\n// updated\n")
	runGitIn(t, source, "add", ".")
	runGitIn(t, source, "commit", "-m", "second commit")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runUpdate(umbrella, cmd, nil); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	headAfter := runGitIn(t, repoDir, "rev-parse", "HEAD")
	if headBefore == headAfter {
		t.Fatalf("expected tracked repo HEAD to change after update, stayed at %s", headBefore)
	}
	assertPathExists(t, filepath.Join(repoDir, "new_feature.go"))
}

func TestRunUpdate_WorkspaceSkipsMissingRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-update-v2-missing")
	if err := os.RemoveAll(repoDir); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runUpdate(umbrella, cmd, nil); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "not cloned") {
		t.Fatalf("expected not-cloned warning, got:\n%s", stderrBuf.String())
	}
}

func TestRunUpdate_WorkspaceInvalidSelector(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-update-v2-invalid")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	err := runUpdate(umbrella, cmd, []string{"ghost-repo"})
	if err == nil {
		t.Fatal("expected error for invalid workspace selector")
	}
	if !strings.Contains(err.Error(), "not a tracked repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}
