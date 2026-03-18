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
// Integration tests — runAdd
// ---------------------------------------------------------------------------

// TestRunAdd_LocalRelativePath verifies that a sibling repo can be added via a
// relative path and that .gitmodules and the submodule directory are created.
func TestRunAdd_LocalRelativePath(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth")
	rel := relPath(umbrella, source)

	if err := runAdd(umbrella, "", &cobra.Command{}, []string{rel}); err != nil {
		t.Fatalf("runAdd with relative path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, ".gitmodules"))
	assertPathExists(t, filepath.Join(umbrella, filepath.Base(source)))
}

// TestRunAdd_LocalRelativePathWithDestination verifies that the optional
// destination [path] argument places the submodule at the requested location.
func TestRunAdd_LocalRelativePathWithDestination(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-billing")
	rel := relPath(umbrella, source)
	dest := "services/billing"

	if err := runAdd(umbrella, "", &cobra.Command{}, []string{rel, dest}); err != nil {
		t.Fatalf("runAdd with destination path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, dest))
	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), dest)
}

// TestRunAdd_LocalAbsolutePath_PrintsWarning verifies that passing an absolute
// local path triggers the portability warning on stderr.
func TestRunAdd_LocalAbsolutePath_PrintsWarning(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-notifications")

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	// We do not assert the error here — we only care about the warning.
	_ = runAdd(umbrella, "", cmd, []string{source})

	if !strings.Contains(stderrBuf.String(), "absolute path") {
		t.Errorf("expected stderr to contain 'absolute path' warning, got:\n%s", stderrBuf.String())
	}
}

// TestRunAdd_LocalAbsolutePath_Succeeds verifies that an absolute local path
// still works (warning is non-fatal).
func TestRunAdd_LocalAbsolutePath_Succeeds(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-gateway")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(umbrella, "", cmd, []string{source}); err != nil {
		t.Errorf("runAdd with absolute path: %v", err)
	}

	assertPathExists(t, filepath.Join(umbrella, ".gitmodules"))
	assertPathExists(t, filepath.Join(umbrella, filepath.Base(source)))
}

// TestRunAdd_WithBranch verifies that the --branch flag is recorded in
// .gitmodules.
func TestRunAdd_WithBranch(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-branch")

	rel := relPath(umbrella, source)
	if err := runAdd(umbrella, "main", &cobra.Command{}, []string{rel}); err != nil {
		t.Fatalf("runAdd with --branch: %v", err)
	}

	assertFileContains(t, filepath.Join(umbrella, ".gitmodules"), "branch = main")
}

// TestRunAdd_NonExistentSource verifies that pointing to a path that does not
// exist results in an error.
func TestRunAdd_NonExistentSource(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := newUmbrellaRepo(t)

	err := runAdd(umbrella, "", &cobra.Command{}, []string{"../repo-that-does-not-exist-xyz"})
	if err == nil {
		t.Error("expected an error for a non-existent source repo, got nil")
	}
}

// TestRunAdd_EmptySourceRepo verifies that adding a repo with no commits
// (which git rejects as a submodule source) results in an error.
func TestRunAdd_EmptySourceRepo(t *testing.T) {
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
	err := runAdd(umbrella, "", &cobra.Command{}, []string{rel})
	if err == nil {
		t.Error("expected an error when adding a repo with no commits, got nil")
	}
}

// TestRunAdd_DuplicateAdd verifies that attempting to add the same submodule
// twice results in an error on the second call.
func TestRunAdd_DuplicateAdd(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-dup")
	rel := relPath(umbrella, source)

	if err := runAdd(umbrella, "", &cobra.Command{}, []string{rel}); err != nil {
		t.Fatalf("first runAdd: %v", err)
	}

	if err := runAdd(umbrella, "", &cobra.Command{}, []string{rel}); err == nil {
		t.Error("expected an error when adding the same submodule twice, got nil")
	}
}
