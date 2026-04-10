package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/spf13/cobra"
)

func TestIsLocalPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"https://github.com/org/repo", false},
		{"https://github.com/org/repo.git", false},
		{"http://github.com/org/repo", false},
		{"git://github.com/org/repo.git", false},
		{"ssh://git@github.com/org/repo", false},
		{"ssh://git@github.com:2222/org/repo", false},
		{"git@github.com:org/repo.git", false},
		{"git@github.com:org/repo", false},
		{"git@gitlab.com:group/subgroup/repo.git", false},
		{"../service-auth", true},
		{"./service-auth", true},
		{"../../some/nested/repo", true},
		{"sibling-repo", true},
		{"services/auth", true},
		{"/home/user/projects/service-auth", true},
		{"/absolute/path/to/repo", true},
		{`C:\\Users\\user\\repos\\service`, true},
		{`..\\service-auth`, true},
		{`.\\service-auth`, true},
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

func TestRunTrack_WorkspaceAddsManifestRepo(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-auth-v2")
	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
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

	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	rel := relPath(umbrella, source)
	if err := runTrack(umbrella, "main", &cobra.Command{}, []string{rel, "services/auth"}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "\"branch\": \"main\"")
	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "services/auth")
}

func TestRunTrack_WorkspaceAbsolutePathPrintsWarning(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-absolute-v2")
	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)
	if err := runTrack(umbrella, "", cmd, []string{source}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "absolute path") {
		t.Fatalf("expected absolute path warning, got:\n%s", stderrBuf.String())
	}
}

func TestRunTrack_WorkspaceNonExistentSource(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := newUmbrellaRepo(t)
	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	err := runTrack(umbrella, "", &cobra.Command{}, []string{"../repo-that-does-not-exist-xyz"})
	if err == nil {
		t.Fatal("expected an error for a non-existent source repo")
	}
}
