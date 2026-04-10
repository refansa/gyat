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

// TestRunInit_EmptyDirectory verifies that gyat init creates both the git
// repository and the .gyat manifest in a fresh directory.
func TestRunInit_EmptyDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)

	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit in empty directory: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, ".git"))
	assertPathExists(t, filepath.Join(dir, manifest.FileName))
	assertPathAbsent(t, filepath.Join(dir, ".gitignore"))
}

// TestRunInit_AlreadyAGitRepoCreatesManifest verifies that running gyat init in
// an existing git repository creates the .gyat manifest if it is missing.
func TestRunInit_AlreadyAGitRepoCreatesManifest(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)

	// Running init a second time should reinitialise cleanly.
	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit on existing repo returned unexpected error: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, manifest.FileName))
}

// TestRunInit_PrintsReadyMessage verifies that a new workspace prints a manifest
// creation message and a follow-up gyat track hint.
func TestRunInit_PrintsReadyMessage(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	var stderrBuf bytes.Buffer
	ic := &cobra.Command{}
	ic.SetErr(&stderrBuf)

	if err := runInit(dir, ic, nil); err != nil {
		t.Errorf("runInit: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "created .gyat manifest") {
		t.Errorf("expected stderr to mention created .gyat manifest, got:\n%s", stderrBuf.String())
	}
	if !strings.Contains(stderrBuf.String(), "gyat track") {
		t.Errorf("expected hint to mention 'gyat track' on stderr, got:\n%s", stderrBuf.String())
	}
}

// TestRunInit_ExistingManifestValidated verifies that an existing manifest is
// loaded successfully and left intact when init is run again.
func TestRunInit_ExistingManifestValidated(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)
	if err := manifest.SaveDir(dir, manifest.File{
		Version: manifest.SupportedVersion,
		Ignore:  []string{"node_modules/"},
		Repos: []manifest.Repo{{
			Name:   "auth",
			Path:   "services/auth",
			URL:    "git@github.com:org/service-auth.git",
			Branch: "main",
		}},
	}); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	var stderrBuf bytes.Buffer
	ic := &cobra.Command{}
	ic.SetErr(&stderrBuf)
	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit with existing manifest: %v", err)
	}

	got, err := manifest.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(got.Repos) != 1 || got.Repos[0].Name != "auth" {
		t.Fatalf("LoadDir repos = %#v, want preserved manifest", got.Repos)
	}
	assertFileContains(t, filepath.Join(dir, ".gitignore"), "# BEGIN gyat managed")
	assertFileContains(t, filepath.Join(dir, ".gitignore"), "/services/auth/")
	if !strings.Contains(stderrBuf.String(), "validated .gyat manifest") {
		t.Errorf("expected stderr to mention validated .gyat manifest, got:\n%s", stderrBuf.String())
	}
	if !strings.Contains(stderrBuf.String(), "updated .gitignore managed block") {
		t.Errorf("expected stderr to mention updated .gitignore managed block, got:\n%s", stderrBuf.String())
	}
}

// TestRunInit_RejectsNestedWorkspace verifies that init refuses to create a
// second workspace under an existing gyat root.
func TestRunInit_RejectsNestedWorkspace(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	root := newUmbrellaRepo(t)
	if err := manifest.SaveDir(root, manifest.Default()); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	nested := filepath.Join(root, "services", "auth")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	ic := &cobra.Command{}
	ic.SetErr(io.Discard)
	err := runInit(nested, ic, nil)
	if err == nil || !strings.Contains(err.Error(), "inside existing workspace") {
		t.Fatalf("runInit error = %v, want nested workspace rejection", err)
	}

	assertPathAbsent(t, filepath.Join(nested, ".git"))
	assertPathAbsent(t, filepath.Join(nested, manifest.FileName))
}

// TestRunInit_DotGitIsDirectory verifies that the .git entry created by init
// is a directory (not a file, which would indicate a git worktree).
func TestRunInit_DotGitIsDirectory(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := t.TempDir()
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)

	if err := runInit(dir, ic, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatalf("stat .git: %v", err)
	}
	if !info.IsDir() {
		t.Error(".git should be a directory for a standard (non-worktree) repository")
	}
}
