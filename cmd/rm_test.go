package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Integration tests — runRm
// ---------------------------------------------------------------------------

func TestRunRm_RootFile(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	// Create and commit a file in the root
	filePath := filepath.Join(dir, "root_file.txt")
	writeFile(t, filePath, "content\n")
	runGitIn(t, dir, "add", "root_file.txt")
	runGitIn(t, dir, "commit", "-m", "add root_file")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	// Remove it
	if err := runRm(dir, false, false, false, cmd, []string{"root_file.txt"}); err != nil {
		t.Fatalf("runRm: %v", err)
	}

	// Verify file is gone from working tree
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("expected root_file.txt to be removed from disk, but it still exists")
	}

	// Verify file is removed from index
	staged := runGitIn(t, dir, "diff", "--cached", "--name-only")
	if !strings.Contains(staged, "root_file.txt") {
		t.Errorf("expected root_file.txt to be staged for deletion\ngit diff --cached:\n%s", staged)
	}
}

func TestRunRm_FileInsideSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-rm-file")
	subDir := filepath.Join(umbrella, subPath)

	// Create and commit a file in the submodule
	filePath := filepath.Join(subDir, "sub_file.txt")
	writeFile(t, filePath, "content\n")
	runGitIn(t, subDir, "add", "sub_file.txt")
	runGitIn(t, subDir, "commit", "-m", "add sub_file")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	// Remove it via umbrella route
	if err := runRm(umbrella, false, false, false, cmd, []string{subPath + "/sub_file.txt"}); err != nil {
		t.Fatalf("runRm: %v", err)
	}

	// Verify file is gone from working tree
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("expected sub_file.txt to be removed from disk, but it still exists")
	}

	// Verify file is removed from index in the submodule
	staged := runGitIn(t, subDir, "diff", "--cached", "--name-only")
	if !strings.Contains(staged, "sub_file.txt") {
		t.Errorf("expected sub_file.txt to be staged for deletion in submodule\ngit diff --cached:\n%s", staged)
	}
}

func TestRunRm_SubmoduleRoot_FailsWithHint(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-rm-root")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	err := runRm(umbrella, false, false, false, cmd, []string{subPath})
	if err == nil {
		t.Fatal("expected runRm to fail when targeting submodule root directly")
	}

	expectedMsg := "to remove submodule '" + subPath + "', use 'gyat untrack " + subPath + "'"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to contain %q, got: %v", expectedMsg, err)
	}
}

func TestRunRm_Cached(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	filePath := filepath.Join(dir, "cached_file.txt")
	writeFile(t, filePath, "content\n")
	runGitIn(t, dir, "add", "cached_file.txt")
	runGitIn(t, dir, "commit", "-m", "add cached_file")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	// Remove it with --cached
	if err := runRm(dir, true, false, false, cmd, []string{"cached_file.txt"}); err != nil {
		t.Fatalf("runRm: %v", err)
	}

	// Verify file STILL exists on disk
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected cached_file.txt to remain on disk, but it was deleted")
	}

	// Verify file is removed from index
	staged := runGitIn(t, dir, "diff", "--cached", "--name-only")
	if !strings.Contains(staged, "cached_file.txt") {
		t.Errorf("expected cached_file.txt to be staged for deletion\ngit diff --cached:\n%s", staged)
	}
}
