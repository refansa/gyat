package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
)

func TestPagerPipePreservesOutputBytes(t *testing.T) {
	binary := buildGyatBinary(t)
	workspace := newIntegrationWorkspace(t)

	stdoutDefault, stderrDefault, exitDefault := runGyat(t, binary, workspace, nil, "status")
	stdoutNoPager, stderrNoPager, exitNoPager := runGyat(t, binary, workspace, nil, "status", "--no-pager")

	if exitDefault != exitNoPager {
		t.Fatalf("exit code mismatch: default=%d no-pager=%d", exitDefault, exitNoPager)
	}
	if !bytes.Equal(stdoutDefault, stdoutNoPager) {
		t.Fatalf("stdout bytes differ\ndefault:\n%s\nno-pager:\n%s", stdoutDefault, stdoutNoPager)
	}
	if !bytes.Equal(stderrDefault, stderrNoPager) {
		t.Fatalf("stderr bytes differ\ndefault:\n%s\nno-pager:\n%s", stderrDefault, stderrNoPager)
	}
}

func buildGyatBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "gyat-test")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = repoRoot(t)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	return binary
}

func newIntegrationWorkspace(t *testing.T) string {
	t.Helper()

	workspace := t.TempDir()
	runGit(t, workspace, "init")
	runGit(t, workspace, "config", "user.email", "test@gyat.test")
	runGit(t, workspace, "config", "user.name", "gyat test")
	runGit(t, workspace, "config", "commit.gpgsign", "false")
	runGit(t, workspace, "config", "core.autocrlf", "false")

	if err := manifest.SaveDir(workspace, manifest.Default()); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	runGit(t, workspace, "add", ".gyat")
	runGit(t, workspace, "commit", "-m", "initial workspace")

	return workspace
}

func runGyat(t *testing.T, binary, workspace string, env []string, args ...string) ([]byte, []byte, int) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Dir = workspace
	cmd.Env = append(os.Environ(), env...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), stderr.Bytes(), 0
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run gyat: %v", err)
	}

	return stdout.Bytes(), stderr.Bytes(), exitErr.ExitCode()
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
