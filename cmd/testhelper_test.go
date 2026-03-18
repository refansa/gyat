package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// skipIfNoGit skips the test if git is not available in PATH.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH, skipping")
	}
}

// newUmbrellaRepo creates a new empty git repository in a fresh temp directory.
// No initial commit is created — gyat add does not require one.
func newUmbrellaRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGitIn(t, dir, "init")
	runGitIn(t, dir, "config", "user.email", "test@gyat.test")
	runGitIn(t, dir, "config", "user.name", "gyat test")
	runGitIn(t, dir, "config", "commit.gpgsign", "false")
	runGitIn(t, dir, "config", "core.autocrlf", "false")

	return dir
}

// newSourceRepo creates a new git repository with an initial commit, making it
// suitable as a submodule source (git submodule add requires a non-empty repo).
func newSourceRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGitIn(t, dir, "init")
	runGitIn(t, dir, "config", "user.email", "test@gyat.test")
	runGitIn(t, dir, "config", "user.name", "gyat test")
	runGitIn(t, dir, "config", "commit.gpgsign", "false")
	runGitIn(t, dir, "config", "core.autocrlf", "false")

	writeFile(t, filepath.Join(dir, "README.md"), "# test source repo\n")
	runGitIn(t, dir, "add", ".")
	runGitIn(t, dir, "commit", "-m", "initial commit")

	return dir
}

// newTestSetup creates an umbrella repo and a source repo as siblings under
// the same parent directory. This layout makes relative paths like
// ../source-repo valid and portable within the test.
//
// Layout:
//
//	<base>/
//	  umbrella/   ← gyat-managed repo
//	  <name>/     ← repo to be added as submodule
func newTestSetup(t *testing.T, sourceName string) (umbrella, source string) {
	t.Helper()

	base := t.TempDir()

	umbrella = filepath.Join(base, "umbrella")
	source = filepath.Join(base, sourceName)

	if err := os.MkdirAll(umbrella, 0o755); err != nil {
		t.Fatalf("create umbrella dir: %v", err)
	}
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}

	// Umbrella — no initial commit needed.
	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	// Source — needs a commit to be usable as a submodule.
	runGitIn(t, source, "init")
	runGitIn(t, source, "config", "user.email", "test@gyat.test")
	runGitIn(t, source, "config", "user.name", "gyat test")
	runGitIn(t, source, "config", "commit.gpgsign", "false")
	runGitIn(t, source, "config", "core.autocrlf", "false")
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")
	runGitIn(t, source, "add", ".")
	runGitIn(t, source, "commit", "-m", "initial commit")

	return umbrella, source
}

// runGitIn runs a git command inside dir and returns trimmed stdout.
// The test is immediately failed if the command exits non-zero.
func runGitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s (in %s) failed: %v\n%s",
			strings.Join(args, " "), dir, err, out)
	}

	return strings.TrimSpace(string(out))
}

// writeFile creates (or overwrites) a file at path with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

// relPath builds a cross-platform relative path from a parent directory to a
// target, using forward slashes so that git always accepts it on Windows too.
func relPath(from, to string) string {
	rel, err := filepath.Rel(from, to)
	if err != nil {
		// Fall back to ../basename which is always correct for siblings.
		rel = "../" + filepath.Base(to)
	}
	// Git accepts forward slashes on all platforms.
	return filepath.ToSlash(rel)
}

// assertPathExists fails the test if path does not exist.
func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected path to exist: %s", path)
	}
}

// assertPathAbsent fails the test if path exists.
func assertPathAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected path to be absent: %s", path)
	}
}

// assertFileContains fails the test if path does not exist or its content does
// not contain substr.
func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("file %s does not contain %q\ncontent:\n%s", path, substr, data)
	}
}

// assertFileNotContains fails the test if path exists and its content contains
// substr. A missing file is treated as not containing substr.
func assertFileNotContains(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return // absent file trivially does not contain the string
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.Contains(string(data), substr) {
		t.Errorf("file %s should not contain %q\ncontent:\n%s", path, substr, data)
	}
}
