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

func TestRunSync_WorkspaceClonesMissingRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-sync-v2-clone")
	if err := os.RemoveAll(repoDir); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runSync(umbrella, cmd, nil); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	assertPathExists(t, repoDir)
	assertPathExists(t, filepath.Join(repoDir, "main.go"))
	if !strings.Contains(stderrBuf.String(), "cloning 'svc-sync-v2-clone'") {
		t.Fatalf("expected clone message, got:\n%s", stderrBuf.String())
	}
}

func TestRunSync_WorkspaceResetsOriginURL(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "svc-sync-v2-remote")

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
	commitWorkspaceMetadata(t, umbrella)

	repoDir := filepath.Join(umbrella, "svc-sync-v2-remote")
	bogusRemote := filepath.Join(filepath.Dir(umbrella), "bogus-remote")
	runGitIn(t, repoDir, "remote", "set-url", "origin", bogusRemote)

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runSync(umbrella, cmd, nil); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	remoteURL := runGitIn(t, repoDir, "config", "--get", "remote.origin.url")
	same, err := samePath(remoteURL, source)
	if err != nil {
		t.Fatalf("samePath: %v", err)
	}
	if !same {
		t.Fatalf("expected origin URL %q, got %q", source, remoteURL)
	}
}

func TestRunSync_WorkspaceInvalidSelector(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-sync-v2-invalid")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	err := runSync(umbrella, cmd, []string{"ghost-repo"})
	if err == nil {
		t.Fatal("expected error for invalid workspace selector")
	}
	if !strings.Contains(err.Error(), "not a tracked repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSync_WithParallelClonesReposInOrder(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDirs := setupTrackedWorkspaceRepos(t, "svc-sync-v2-parallel-a", "svc-sync-v2-parallel-b")
	for _, repoDir := range repoDirs {
		if err := os.RemoveAll(repoDir); err != nil {
			t.Fatalf("RemoveAll: %v", err)
		}
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runSyncWithFlagsFrom(umbrella, umbrella, workspaceTargetFlags{parallel: true}, cmd, nil); err != nil {
		t.Fatalf("runSyncWithFlagsFrom: %v", err)
	}

	assertPathExists(t, repoDirs["svc-sync-v2-parallel-a"])
	assertPathExists(t, repoDirs["svc-sync-v2-parallel-b"])
	errOut := stderrBuf.String()
	first := strings.Index(errOut, "cloning 'svc-sync-v2-parallel-a'")
	second := strings.Index(errOut, "cloning 'svc-sync-v2-parallel-b'")
	if first == -1 || second == -1 || first > second {
		t.Fatalf("expected ordered clone messages, got:\n%s", errOut)
	}
}
