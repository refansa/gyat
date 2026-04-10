package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/spf13/cobra"
)

func TestRunExec_AllTargets(t *testing.T) {
	root := newExecWorkspace(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&stdout)
	command.SetErr(&stderr)

	err := runExec(root, nil, nil, false, false, false, command, []string{os.Args[0], "-test.run=TestExecHelperProcess", "--", "print-base"})
	if err != nil {
		t.Fatalf("runExec: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "umbrella repository") {
		t.Fatalf("expected umbrella repository in output, got:\n%s", output)
	}
	if !strings.Contains(output, "services/auth") || !strings.Contains(output, "services/billing") {
		t.Fatalf("expected repo labels in output, got:\n%s", output)
	}
	if !strings.Contains(stderr.String(), "running '") {
		t.Fatalf("expected progress message on stderr, got:\n%s", stderr.String())
	}
}

func TestRunExec_RepoSelectorWithoutRoot(t *testing.T) {
	root := newExecWorkspace(t)

	var stdout bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&stdout)
	command.SetErr(new(bytes.Buffer))

	err := runExec(root, []string{"auth"}, nil, true, false, false, command, []string{os.Args[0], "-test.run=TestExecHelperProcess", "--", "print-base"})
	if err != nil {
		t.Fatalf("runExec: %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "auth" {
		t.Fatalf("stdout = %q, want %q", got, "auth")
	}
}

func TestRunExec_ContinueOnError(t *testing.T) {
	root := newExecWorkspace(t)

	var stdout bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&stdout)
	command.SetErr(new(bytes.Buffer))

	err := runExec(root, nil, nil, true, false, true, command, []string{os.Args[0], "-test.run=TestExecHelperProcess", "--", "fail-on-dir", "billing"})
	if err == nil || !strings.Contains(err.Error(), "command failed in 1 target") {
		t.Fatalf("runExec error = %v, want aggregate failure", err)
	}
	if !strings.Contains(stdout.String(), "auth") || !strings.Contains(stdout.String(), "billing") {
		t.Fatalf("expected output for all selected repos, got:\n%s", stdout.String())
	}
}

func TestExecHelperProcess(t *testing.T) {
	if len(os.Args) < 4 || os.Args[1] != "-test.run=TestExecHelperProcess" {
		return
	}

	args := os.Args[3:]
	mode := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		os.Exit(2)
	}

	switch mode {
	case "print-base":
		_, _ = os.Stdout.WriteString(filepath.Base(cwd))
		os.Exit(0)
	case "fail-on-dir":
		if len(args) < 2 {
			os.Exit(3)
		}
		if filepath.Base(cwd) == args[1] {
			_, _ = os.Stderr.WriteString("forced failure")
			os.Exit(7)
		}
		_, _ = os.Stdout.WriteString(filepath.Base(cwd))
		os.Exit(0)
	default:
		os.Exit(4)
	}
}

func newExecWorkspace(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "services", "auth"),
		filepath.Join(root, "services", "billing"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
	}

	if err := manifest.SaveDir(root, manifest.File{
		Version: manifest.SupportedVersion,
		Repos: []manifest.Repo{
			{Name: "auth", Path: "services/auth", URL: "git@github.com:org/auth.git", Groups: []string{"backend"}},
			{Name: "billing", Path: "services/billing", URL: "git@github.com:org/billing.git", Groups: []string{"payments"}},
		},
	}); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	return root
}
