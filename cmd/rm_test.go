package cmd

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunRm_WorkspaceRootFile(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-rm-v2-root-file")
	filePath := filepath.Join(umbrella, "root_file.txt")
	writeFile(t, filePath, "content\n")
	runGitIn(t, umbrella, "add", "root_file.txt")
	runGitIn(t, umbrella, "commit", "-m", "add root file")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runRm(umbrella, false, false, false, cmd, []string{"root_file.txt"}); err != nil {
		t.Fatalf("runRm root file: %v", err)
	}

	assertPathAbsent(t, filePath)
	if staged := stagedFilesInDir(t, umbrella); !strings.Contains(staged, "root_file.txt") {
		t.Fatalf("expected root_file.txt to be staged for deletion\ngit diff --cached:\n%s", staged)
	}
}

func TestRunRm_WorkspaceFileInsideRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDir := setupTrackedWorkspaceRepo(t, "svc-rm-v2-repo-file")
	filePath := filepath.Join(repoDir, "repo_file.txt")
	writeFile(t, filePath, "content\n")
	runGitIn(t, repoDir, "add", "repo_file.txt")
	runGitIn(t, repoDir, "commit", "-m", "add repo file")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runRm(umbrella, false, false, false, cmd, []string{"svc-rm-v2-repo-file/repo_file.txt"}); err != nil {
		t.Fatalf("runRm repo file: %v", err)
	}

	assertPathAbsent(t, filePath)
	if staged := stagedFilesInDir(t, repoDir); !strings.Contains(staged, "repo_file.txt") {
		t.Fatalf("expected repo_file.txt to be staged for deletion\ngit diff --cached:\n%s", staged)
	}
}

func TestRunRm_WorkspaceRepoRootFailsWithHint(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, _ := setupTrackedWorkspaceRepo(t, "svc-rm-v2-root")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	err := runRm(umbrella, false, false, false, cmd, []string{"svc-rm-v2-root"})
	if err == nil {
		t.Fatal("expected error when removing tracked repository root")
	}
	if !strings.Contains(err.Error(), "use 'gyat untrack svc-rm-v2-root'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunRm_WithRepoFlagRemovesOnlySelectedRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, repoDirs := setupTrackedWorkspaceRepos(t, "svc-rm-v2-flag-a", "svc-rm-v2-flag-b")
	for _, repoDir := range repoDirs {
		filePath := filepath.Join(repoDir, "generated.txt")
		writeFile(t, filePath, "content\n")
		runGitIn(t, repoDir, "add", "generated.txt")
		runGitIn(t, repoDir, "commit", "-m", "add generated file")
	}

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	flags := workspaceTargetFlags{
		repoSelectors: []string{"svc-rm-v2-flag-a"},
		noRoot:        true,
	}
	if err := runRmWithFlagsFrom(umbrella, umbrella, flags, false, false, false, cmd, []string{"generated.txt"}); err != nil {
		t.Fatalf("runRmWithFlagsFrom: %v", err)
	}

	assertPathAbsent(t, filepath.Join(repoDirs["svc-rm-v2-flag-a"], "generated.txt"))
	assertPathExists(t, filepath.Join(repoDirs["svc-rm-v2-flag-b"], "generated.txt"))
	if staged := stagedFilesInDir(t, repoDirs["svc-rm-v2-flag-a"]); !strings.Contains(staged, "generated.txt") {
		t.Fatalf("expected selected repo to stage generated.txt for deletion\ngit diff --cached:\n%s", staged)
	}
	if staged := stagedFilesInDir(t, repoDirs["svc-rm-v2-flag-b"]); staged != "" {
		t.Fatalf("expected unselected repo to remain untouched, got: %q", staged)
	}
}
