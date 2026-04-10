package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/spf13/cobra"
)

func setupTrackedWorkspaceRepo(t *testing.T, repoName string) (umbrella, repoDir string) {
	t.Helper()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, repoName)

	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("setupTrackedWorkspaceRepo: runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(io.Discard)
	if err := runTrack(umbrella, "", trackCmd, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("setupTrackedWorkspaceRepo: runTrack: %v", err)
	}

	commitWorkspaceMetadata(t, umbrella)
	return umbrella, filepath.Join(umbrella, repoName)
}

func commitWorkspaceMetadata(t *testing.T, umbrella string) {
	t.Helper()
	runGitIn(t, umbrella, "add", manifest.FileName, ".gitignore")
	runGitIn(t, umbrella, "commit", "-m", "track workspace metadata")
}

func dirtyTrackedRepo(t *testing.T, repoDir string) {
	t.Helper()
	writeFile(t, filepath.Join(repoDir, "new_feature.go"), "package main\n// new feature\n")
}

func stagedFilesInDir(t *testing.T, dir string) string {
	t.Helper()
	return runGitIn(t, dir, "diff", "--cached", "--name-only")
}

func TestHasWorkingTreeChanges_Empty(t *testing.T) {
	t.Parallel()
	if hasWorkingTreeChanges("") {
		t.Error("expected false for empty status output")
	}
}

func TestHasWorkingTreeChanges_StagedOnly(t *testing.T) {
	t.Parallel()
	out := "A  README.md\nA  docs/guide.md\n"
	if hasWorkingTreeChanges(out) {
		t.Errorf("expected false for staged-only output\ninput: %q", out)
	}
}

func TestHasWorkingTreeChanges_Untracked(t *testing.T) {
	t.Parallel()
	out := "?? new_file.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for untracked file\ninput: %q", out)
	}
}

func TestHasWorkingTreeChanges_ModifiedInWorkingTree(t *testing.T) {
	t.Parallel()
	out := " M main.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for working-tree modification\ninput: %q", out)
	}
}

func TestHasWorkingTreeChanges_DeletedInWorkingTree(t *testing.T) {
	t.Parallel()
	out := " D old_file.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for working-tree deletion\ninput: %q", out)
	}
}

func TestHasWorkingTreeChanges_StagedAndModified(t *testing.T) {
	t.Parallel()
	out := "AM handler.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for staged-and-modified file\ninput: %q", out)
	}
}

func TestHasWorkingTreeChanges_MixedStagedAndUntracked(t *testing.T) {
	t.Parallel()
	out := "A  README.md\n?? scratch.txt\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true when output includes an untracked file\ninput: %q", out)
	}
}

func TestRunAdd_WorkspaceNothingToStage(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-v2-clean")

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd workspace: %v", err)
	}
	if !strings.Contains(stderrBuf.String(), "nothing to stage") {
		t.Fatalf("expected nothing-to-stage message, got:\n%s", stderrBuf.String())
	}
}

func TestRunAdd_WorkspaceStagesRootAndRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-v2-stage-all")

	writeFile(t, filepath.Join(umbrella, ".editorconfig"), "root = true\n")
	dirtyTrackedRepo(t, repoDir)

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd workspace: %v", err)
	}

	if staged := stagedFilesInDir(t, umbrella); !strings.Contains(staged, ".editorconfig") {
		t.Errorf("expected .editorconfig to be staged in umbrella root\ngit diff --cached:\n%s", staged)
	}
	if staged := stagedFilesInDir(t, repoDir); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected new_feature.go to be staged in tracked repo\ngit diff --cached:\n%s", staged)
	}
}

func TestRunAdd_WorkspaceStagesSpecificRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-v2-a")
	svcB := filepath.Join(base, "svc-v2-b")

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
	if err := runTrack(umbrella, "", trackCmd, []string{"../svc-v2-a"}); err != nil {
		t.Fatalf("track svc-v2-a: %v", err)
	}
	if err := runTrack(umbrella, "", trackCmd, []string{"../svc-v2-b"}); err != nil {
		t.Fatalf("track svc-v2-b: %v", err)
	}
	commitWorkspaceMetadata(t, umbrella)

	dirtyTrackedRepo(t, filepath.Join(umbrella, "svc-v2-a"))
	dirtyTrackedRepo(t, filepath.Join(umbrella, "svc-v2-b"))

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	if err := runAdd(umbrella, cmd, []string{"svc-v2-a"}); err != nil {
		t.Fatalf("runAdd svc-v2-a: %v", err)
	}

	if staged := stagedFilesInDir(t, filepath.Join(umbrella, "svc-v2-a")); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected svc-v2-a to have staged changes\ngot: %q", staged)
	}
	if staged := stagedFilesInDir(t, filepath.Join(umbrella, "svc-v2-b")); staged != "" {
		t.Errorf("expected svc-v2-b to have no staged changes, got: %q", staged)
	}
}

func TestRunAdd_WorkspaceFileInsideRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-v2-file-route")

	writeFile(t, filepath.Join(repoDir, "handler.go"), "package main\n")
	writeFile(t, filepath.Join(repoDir, "other.go"), "package main\n")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(umbrella, cmd, []string{"svc-v2-file-route/handler.go"}); err != nil {
		t.Fatalf("runAdd workspace file path: %v", err)
	}

	staged := stagedFilesInDir(t, repoDir)
	if !strings.Contains(staged, "handler.go") {
		t.Errorf("expected handler.go to be staged\ngit diff --cached:\n%s", staged)
	}
	if strings.Contains(staged, "other.go") {
		t.Errorf("expected other.go to NOT be staged\ngit diff --cached:\n%s", staged)
	}
}

func TestRunAdd_WorkspaceDotUsesCurrentDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-v2-current-dir")

	writeFile(t, filepath.Join(umbrella, ".editorconfig"), "root = true\n")
	writeFile(t, filepath.Join(repoDir, "handler.go"), "package main\n")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAddFrom(repoDir, umbrella, cmd, []string{"."}); err != nil {
		t.Fatalf("runAddFrom current repo: %v", err)
	}

	if staged := stagedFilesInDir(t, repoDir); !strings.Contains(staged, "handler.go") {
		t.Errorf("expected handler.go to be staged from current repo dir\ngit diff --cached:\n%s", staged)
	}
	if staged := stagedFilesInDir(t, umbrella); staged != "" {
		t.Errorf("expected umbrella root to be untouched, got staged: %q", staged)
	}
}

func TestRunAdd_WithRepoFlagStagesOnlySelectedRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDirs := setupTrackedWorkspaceRepos(t, "svc-v2-flag-a", "svc-v2-flag-b")
	writeFile(t, filepath.Join(umbrella, ".editorconfig"), "root = true\n")
	dirtyTrackedRepo(t, repoDirs["svc-v2-flag-a"])
	dirtyTrackedRepo(t, repoDirs["svc-v2-flag-b"])

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	flags := workspaceTargetFlags{
		repoSelectors: []string{"svc-v2-flag-a"},
		noRoot:        true,
	}
	if err := runAddWithFlagsFrom(umbrella, umbrella, flags, cmd, nil); err != nil {
		t.Fatalf("runAddWithFlagsFrom: %v", err)
	}

	if staged := stagedFilesInDir(t, repoDirs["svc-v2-flag-a"]); !strings.Contains(staged, "new_feature.go") {
		t.Fatalf("expected selected repo to stage new_feature.go\ngit diff --cached:\n%s", staged)
	}
	if staged := stagedFilesInDir(t, repoDirs["svc-v2-flag-b"]); staged != "" {
		t.Fatalf("expected unselected repo to remain unstaged, got: %q", staged)
	}
	if staged := stagedFilesInDir(t, umbrella); staged != "" {
		t.Fatalf("expected umbrella root to remain unstaged, got: %q", staged)
	}
}
