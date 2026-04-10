package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunPull_WorkspaceSkipsLocalPathRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-pull-v2-local")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPull(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPull: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "hint:") || !strings.Contains(stderr, "svc-pull-v2-local") {
		t.Fatalf("expected local-path hint for tracked repo, got:\n%s", stderr)
	}
}

func TestRunPull_WorkspaceInvalidSelector(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-pull-v2-invalid")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)

	err := runPull(umbrella, false, cc, []string{"not-a-repo"})
	if err == nil {
		t.Fatal("expected error for invalid workspace selector")
	}
	if !strings.Contains(err.Error(), "not a tracked repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPull_WorkspaceDotUsesCurrentDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-pull-v2-dot")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPullFrom(repoDir, umbrella, false, cc, []string{"."}); err != nil {
		t.Fatalf("runPullFrom: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "svc-pull-v2-dot") || !strings.Contains(stderr, "hint:") {
		t.Fatalf("expected current-directory repo to be selected, got:\n%s", stderr)
	}
	if strings.Contains(stderr, "umbrella repository") {
		t.Fatalf("unexpected umbrella-specific output in stderr:\n%s", stderr)
	}
}
