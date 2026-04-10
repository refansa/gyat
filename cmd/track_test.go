package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/internal/manifest"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Unit tests — isLocalPath
// ---------------------------------------------------------------------------

func TestIsLocalPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		// Remote URLs — must return false
		{"https://github.com/org/repo", false},
		{"https://github.com/org/repo.git", false},
		{"http://github.com/org/repo", false},
		{"git://github.com/org/repo.git", false},
		{"ssh://git@github.com/org/repo", false},
		{"ssh://git@github.com:2222/org/repo", false},
		{"git@github.com:org/repo.git", false},
		{"git@github.com:org/repo", false},
		{"git@gitlab.com:group/subgroup/repo.git", false},

		// Local relative paths — must return true
		{"../service-auth", true},
		{"./service-auth", true},
		{"../../some/nested/repo", true},
		{"sibling-repo", true},
		{"services/auth", true},

		// Local absolute paths — must return true
		{"/home/user/projects/service-auth", true},
		{"/absolute/path/to/repo", true},

		// Windows-style paths — must return true
		{`C:\Users\user\repos\service`, true},
		{`..\service-auth`, true},
		{`.\service-auth`, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := isLocalPath(tt.input)
			if got != tt.want {
				t.Errorf("isLocalPath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Integration tests — runTrack
// ---------------------------------------------------------------------------

// TestRunTrack_LocalRelativePath verifies that a sibling repo can be tracked via a
// relative path and that .gitmodules and the submodule directory are created.
func TestRunTrack_LocalRelativePath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth")
	rel := relPath(umbrella, source)

	if err := runTrack(umbrella, "", &cobra.Command{}, []string{rel}); err != nil {
		t.Fatalf("runTrack with relative path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, ".gitmodules"))
	assertPathExists(t, filepath.Join(umbrella, filepath.Base(source)))
}

// TestRunTrack_LocalRelativePathWithDestination verifies that the optional
// destination [path] argument places the submodule at the requested location.
func TestRunTrack_LocalRelativePathWithDestination(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-billing")
	rel := relPath(umbrella, source)
	dest := "services/billing"

	if err := runTrack(umbrella, "", &cobra.Command{}, []string{rel, dest}); err != nil {
		t.Fatalf("runTrack with destination path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, dest))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), dest)
}

// TestRunTrack_LocalAbsolutePath_PrintsWarning verifies that passing an absolute
// local path triggers the portability warning on stderr.
func TestRunTrack_LocalAbsolutePath_PrintsWarning(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-notifications")

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	// We do not assert the error here — we only care about the warning.
	_ = runTrack(umbrella, "", cmd, []string{source})

	if !strings.Contains(stderrBuf.String(), "absolute path") {
		t.Errorf("expected stderr to contain 'absolute path' warning, got:\n%s", stderrBuf.String())
	}
}

// TestRunTrack_LocalAbsolutePath_Succeeds verifies that an absolute local path
// still works (warning is non-fatal).
func TestRunTrack_LocalAbsolutePath_Succeeds(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-gateway")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runTrack(umbrella, "", cmd, []string{source}); err != nil {
		t.Errorf("runTrack with absolute path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, ".gitmodules"))
	assertPathExists(t, filepath.Join(umbrella, filepath.Base(source)))
}

// TestRunTrack_WithBranch verifies that the --branch flag is recorded in
// .gitmodules.
func TestRunTrack_WithBranch(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-branch")

	rel := relPath(umbrella, source)
	if err := runTrack(umbrella, "main", &cobra.Command{}, []string{rel}); err != nil {
		t.Fatalf("runTrack with --branch: %v", err)
	}

	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "branch = main")
}

// TestRunTrack_NonExistentSource verifies that pointing to a path that does not
// exist results in an error.
func TestRunTrack_NonExistentSource(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := newUmbrellaRepo(t)

	err := runTrack(umbrella, "", &cobra.Command{}, []string{"../repo-that-does-not-exist-xyz"})
	if err == nil {
		t.Error("expected an error for a non-existent source repo, got nil")
	}
}

// TestRunTrack_EmptySourceRepo verifies that tracking a repo with no commits
// (which git rejects as a submodule source) results in an error.
func TestRunTrack_EmptySourceRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()

	umbrella := filepath.Join(base, "umbrella")
	empty := filepath.Join(base, "empty-repo")

	if err := os.MkdirAll(umbrella, 0o755); err != nil {
		t.Fatalf("create umbrella: %v", err)
	}
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatalf("create empty: %v", err)
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, empty, "init")

	rel := relPath(umbrella, empty)
	err := runTrack(umbrella, "", &cobra.Command{}, []string{rel})
	if err == nil {
		t.Error("expected an error when tracking a repo with no commits, got nil")
	}
}

// TestRunTrack_DuplicateTrack verifies that attempting to track the same submodule
// twice results in an error on the second call.
func TestRunTrack_DuplicateTrack(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-dup")
	rel := relPath(umbrella, source)

	if err := runTrack(umbrella, "", &cobra.Command{}, []string{rel}); err != nil {
		t.Fatalf("first runTrack: %v", err)
	}

	if err := runTrack(umbrella, "", &cobra.Command{}, []string{rel}); err == nil {
		t.Error("expected an error when tracking the same submodule twice, got nil")
	}
}

func TestRunTrack_WorkspaceAddsManifestRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-v2")
	ic := &cobra.Command{}
	ic.SetErr(io.Discard)
	if err := runInit(umbrella, ic, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	rel := relPath(umbrella, source)
	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)
	if err := runTrack(umbrella, "", cmd, []string{rel}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, "service-auth-v2"))
	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "service-auth-v2")
	assertFileContains(t, filepath.Join(umbrella, ".gitignore"), "/service-auth-v2/")
	if !strings.Contains(stderrBuf.String(), "tracked repository 'service-auth-v2'") {
		t.Fatalf("expected tracked repository message, got:\n%s", stderrBuf.String())
	}
}

func TestRunTrack_WorkspaceWithBranchRecordsManifest(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-main")
	runGitIn(t, source, "branch", "-m", "main")

	ic := &cobra.Command{}
	ic.SetErr(io.Discard)
	if err := runInit(umbrella, ic, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	rel := relPath(umbrella, source)
	if err := runTrack(umbrella, "main", &cobra.Command{}, []string{rel, "services/auth"}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "\"branch\": \"main\"")
	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "services/auth")
}
