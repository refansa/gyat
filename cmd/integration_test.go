package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, "my-service")

	initCmd := &cobra.Command{}
	initCmd.SetErr(io.Discard)
	if err := runInit(umbrella, initCmd, nil); err != nil {
		t.Fatalf("step 1 (gyat init): %v", err)
	}
	assertPathExists(t, filepath.Join(umbrella, ".git"))

	rel := relPath(umbrella, source)
	trackCmd := &cobra.Command{}
	trackCmd.SetErr(io.Discard)
	if err := runTrack(umbrella, "", trackCmd, []string{rel}); err != nil {
		t.Fatalf("step 2 (gyat track %s): %v", rel, err)
	}
	assertPathExists(t, filepath.Join(umbrella, "my-service"))
	assertFileContains(t, filepath.Join(umbrella, ".gyat"), "my-service")
	assertFileContains(t, filepath.Join(umbrella, ".gitignore"), "/my-service/")

	var listOut bytes.Buffer
	listCmd := &cobra.Command{}
	listCmd.SetOut(&listOut)
	listCmd.SetErr(new(bytes.Buffer))
	if err := runList(umbrella, listCmd, nil); err != nil {
		t.Fatalf("step 3 (gyat list): %v", err)
	}

	out := listOut.String()
	if !strings.Contains(out, "my-service") {
		t.Fatalf("step 3: expected 'my-service' in list output\ngot:\n%s", out)
	}
	if !strings.Contains(out, "PATH") {
		t.Fatalf("step 3: expected table header in list output\ngot:\n%s", out)
	}

	untrackCmd := &cobra.Command{}
	untrackCmd.SetErr(io.Discard)
	if err := runUntrack(umbrella, untrackCmd, []string{"my-service"}); err != nil {
		t.Fatalf("step 4 (gyat untrack my-service): %v", err)
	}
	assertPathAbsent(t, filepath.Join(umbrella, "my-service"))
	assertFileNotContains(t, filepath.Join(umbrella, ".gyat"), "my-service")
	assertFileNotContains(t, filepath.Join(umbrella, ".gitignore"), "/my-service/")
}
