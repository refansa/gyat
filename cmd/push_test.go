package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunPush_WorkspaceSkipsLocalPathRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-push-v2-local")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPush(umbrella, false, cc, nil); err != nil {
		t.Fatalf("runPush: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "hint:") || !strings.Contains(stderr, "svc-push-v2-local") {
		t.Fatalf("expected local-path hint for tracked repo, got:\n%s", stderr)
	}
}

func TestRunPush_WorkspaceInvalidSelector(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-push-v2-invalid")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)

	err := runPush(umbrella, false, cc, []string{"not-a-repo"})
	if err == nil {
		t.Fatal("expected error for invalid workspace selector")
	}
	if !strings.Contains(err.Error(), "not a tracked repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPush_WorkspaceDotUsesCurrentDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-push-v2-dot")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runPushFrom(repoDir, umbrella, false, cc, []string{"."}); err != nil {
		t.Fatalf("runPushFrom: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "svc-push-v2-dot") || !strings.Contains(stderr, "hint:") {
		t.Fatalf("expected current-directory repo to be selected, got:\n%s", stderr)
	}
	if strings.Contains(stderr, "umbrella repository") {
		t.Fatalf("unexpected umbrella-specific output in stderr:\n%s", stderr)
	}
}
