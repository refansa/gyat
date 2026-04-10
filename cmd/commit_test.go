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

func TestRunCommit_NoVerifyFlag(t *testing.T) {
	t.Parallel()

	args := buildCommitArgs("test message", true)
	found := false
	for _, arg := range args {
		if arg == "--no-verify" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected --no-verify in commit args\ngot: %v", args)
	}

	argsWithout := buildCommitArgs("test message", false)
	for _, arg := range argsWithout {
		if arg == "--no-verify" {
			t.Fatalf("did not expect --no-verify when noVerify is false\ngot: %v", argsWithout)
		}
	}
}

func TestRunCommit_WorkspaceNothingToCommit(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-commit-v2-nothing")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "should not commit", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "nothing to commit") {
		t.Fatalf("expected stderr to contain 'nothing to commit'\ngot: %s", stderrBuf.String())
	}
}

func TestRunCommit_WorkspaceCommitsRepoAndRoot(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-commit-v2-all")

	writeFile(t, filepath.Join(repoDir, "handler.go"), "package main\n")
	runGitIn(t, repoDir, "add", "handler.go")

	writeFile(t, filepath.Join(umbrella, ".editorconfig"), "root = true\n")
	runGitIn(t, umbrella, "add", ".editorconfig")

	var stderrBuf bytes.Buffer
	cc := &cobra.Command{}
	cc.SetErr(&stderrBuf)

	if err := runCommit(umbrella, "feat: workspace commit", false, cc, nil); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	repoLog := runGitIn(t, repoDir, "log", "--oneline", "-1")
	if !strings.Contains(repoLog, "feat: workspace commit") {
		t.Fatalf("expected commit message in tracked repo log\ngot: %s", repoLog)
	}

	umbrellaLog := runGitIn(t, umbrella, "log", "--oneline", "-1")
	if !strings.Contains(umbrellaLog, "feat: workspace commit") {
		t.Fatalf("expected commit message in umbrella log\ngot: %s", umbrellaLog)
	}
}

func TestRunCommit_WorkspaceTargetedRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-commit-v2-a")
	svcB := filepath.Join(base, "svc-commit-v2-b")

	for _, dir := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, dir := range []string{svcA, svcB} {
		runGitIn(t, dir, "init")
		runGitIn(t, dir, "config", "user.email", "test@gyat.test")
		runGitIn(t, dir, "config", "user.name", "gyat test")
		runGitIn(t, dir, "config", "commit.gpgsign", "false")
		runGitIn(t, dir, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(dir, "main.go"), "package main\n")
		runGitIn(t, dir, "add", ".")
		runGitIn(t, dir, "commit", "-m", "init")
	}

	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(io.Discard)
	if err := runTrack(umbrella, "", trackCmd, []string{"../svc-commit-v2-a"}); err != nil {
		t.Fatalf("track svc-commit-v2-a: %v", err)
	}
	if err := runTrack(umbrella, "", trackCmd, []string{"../svc-commit-v2-b"}); err != nil {
		t.Fatalf("track svc-commit-v2-b: %v", err)
	}
	commitWorkspaceMetadata(t, umbrella)

	repoDirA := filepath.Join(umbrella, "svc-commit-v2-a")
	repoDirB := filepath.Join(umbrella, "svc-commit-v2-b")
	writeFile(t, filepath.Join(repoDirA, "a.go"), "package a\n")
	runGitIn(t, repoDirA, "add", "a.go")
	writeFile(t, filepath.Join(repoDirB, "b.go"), "package b\n")
	runGitIn(t, repoDirB, "add", "b.go")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)
	if err := runCommit(umbrella, "feat: targeted workspace commit", false, cc, []string{"svc-commit-v2-a"}); err != nil {
		t.Fatalf("runCommit: %v", err)
	}

	logA := runGitIn(t, repoDirA, "log", "--oneline", "-1")
	if !strings.Contains(logA, "feat: targeted workspace commit") {
		t.Fatalf("expected commit message in svc-commit-v2-a log\ngot: %s", logA)
	}

	logB := runGitIn(t, repoDirB, "log", "--oneline", "-1")
	if strings.Contains(logB, "feat: targeted workspace commit") {
		t.Fatalf("svc-commit-v2-b should not have been committed\ngot: %s", logB)
	}
}

func TestRunCommit_WorkspaceDotUsesCurrentDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-commit-v2-dot")

	writeFile(t, filepath.Join(repoDir, "handler.go"), "package main\n")
	runGitIn(t, repoDir, "add", "handler.go")
	writeFile(t, filepath.Join(umbrella, ".editorconfig"), "root = true\n")
	runGitIn(t, umbrella, "add", ".editorconfig")
	rootHeadBefore := runGitIn(t, umbrella, "rev-parse", "HEAD")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)
	if err := runCommitFrom(repoDir, umbrella, "feat: repo only", false, cc, []string{"."}); err != nil {
		t.Fatalf("runCommitFrom: %v", err)
	}

	repoLog := runGitIn(t, repoDir, "log", "--oneline", "-1")
	if !strings.Contains(repoLog, "feat: repo only") {
		t.Fatalf("expected commit message in repo log\ngot: %s", repoLog)
	}

	rootHeadAfter := runGitIn(t, umbrella, "rev-parse", "HEAD")
	if rootHeadBefore != rootHeadAfter {
		t.Fatalf("expected umbrella HEAD to remain unchanged\nbefore: %s\nafter:  %s", rootHeadBefore, rootHeadAfter)
	}

	rootStatus := runGitIn(t, umbrella, "status", "--porcelain")
	if !strings.Contains(rootStatus, ".editorconfig") {
		t.Fatalf("expected umbrella root change to remain staged\ngot: %s", rootStatus)
	}
}

func TestRunCommit_WorkspaceInvalidSelector(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-commit-v2-invalid")

	cc := &cobra.Command{}
	cc.SetErr(io.Discard)
	err := runCommit(umbrella, "bad path", false, cc, []string{"not-a-repo"})
	if err == nil {
		t.Fatal("expected error for invalid workspace selector")
	}
	if !strings.Contains(err.Error(), "not a tracked repository or workspace path") {
		t.Fatalf("unexpected error: %v", err)
	}
}
