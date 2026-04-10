package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/spf13/cobra"
)

func TestRunList_WorkspaceNoRepos(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella := t.TempDir()
	initCmd := &cobra.Command{}
	initCmd.SetErr(new(bytes.Buffer))
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var stdout, stderr bytes.Buffer
	listCmd := &cobra.Command{}
	listCmd.SetOut(&stdout)
	listCmd.SetErr(&stderr)
	if err := runList(umbrella, listCmd, nil); err != nil {
		t.Fatalf("runList: %v", err)
	}

	if !strings.Contains(stdout.String(), "no managed repos found") {
		t.Fatalf("expected empty-workspace message, got:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "gyat track") {
		t.Fatalf("expected gyat track hint, got:\n%s", stderr.String())
	}
}

func TestRunList_WorkspaceManifest(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "service-search-v2")
	initCmd := &cobra.Command{}
	initCmd.SetErr(new(bytes.Buffer))
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	trackCmd := &cobra.Command{}
	trackCmd.SetErr(new(bytes.Buffer))
	if err := runTrack(umbrella, "", trackCmd, []string{relPath(umbrella, source)}); err != nil {
		t.Fatalf("runTrack: %v", err)
	}

	var stdout bytes.Buffer
	listCmd := &cobra.Command{}
	listCmd.SetOut(&stdout)
	listCmd.SetErr(new(bytes.Buffer))
	if err := runList(umbrella, listCmd, nil); err != nil {
		t.Fatalf("runList: %v", err)
	}

	if !strings.Contains(stdout.String(), "service-search-v2") {
		t.Fatalf("expected tracked repo path in output, got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "clean") {
		t.Fatalf("expected clean status in output, got:\n%s", stdout.String())
	}
	assertFileContains(t, filepath.Join(umbrella, manifest.FileName), "service-search-v2")
}
