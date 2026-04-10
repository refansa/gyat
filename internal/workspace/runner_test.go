package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCommandCapturesOutputPerTarget(t *testing.T) {
	root := t.TempDir()
	alpha := filepath.Join(root, "alpha")
	beta := filepath.Join(root, "beta")
	for _, dir := range []string{alpha, beta} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	targets := []Target{
		{Label: "alpha", Dir: alpha},
		{Label: "beta", Dir: beta},
	}

	results, err := RunCommand(targets, Command{
		Name: os.Args[0],
		Args: []string{"-test.run=TestWorkspaceHelperProcess", "--", "print-base"},
	}, RunOptions{})
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("RunCommand len = %d, want 2", len(results))
	}
	if results[0].Output != "alpha" || results[1].Output != "beta" {
		t.Fatalf("RunCommand outputs = %#v, want alpha/beta", results)
	}
}

func TestRunCommandContinueOnError(t *testing.T) {
	root := t.TempDir()
	okDir := filepath.Join(root, "ok")
	failDir := filepath.Join(root, "fail")
	for _, dir := range []string{okDir, failDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	results, err := RunCommand([]Target{
		{Label: "ok", Dir: okDir},
		{Label: "fail", Dir: failDir},
	}, Command{
		Name: os.Args[0],
		Args: []string{"-test.run=TestWorkspaceHelperProcess", "--", "fail-on-dir", "fail"},
	}, RunOptions{ContinueOnError: true})
	if err == nil || !strings.Contains(err.Error(), "command failed in 1 target") {
		t.Fatalf("RunCommand error = %v, want aggregate failure", err)
	}
	if len(results) != 2 {
		t.Fatalf("RunCommand len = %d, want 2", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("results[0].Err = %v, want nil", results[0].Err)
	}
	if results[1].Err == nil {
		t.Fatal("results[1].Err = nil, want failure")
	}
}

func TestRunTargetsPreservesOrderInParallel(t *testing.T) {
	t.Parallel()

	targets := []Target{{Label: "slow"}, {Label: "fast"}, {Label: "mid"}}
	results, err := RunTargets(targets, RunOptions{Parallel: true}, func(target Target) (string, error) {
		switch target.Label {
		case "slow":
			time.Sleep(20 * time.Millisecond)
		case "mid":
			time.Sleep(10 * time.Millisecond)
		}
		return target.Label, nil
	})
	if err != nil {
		t.Fatalf("RunTargets: %v", err)
	}

	want := []string{"slow", "fast", "mid"}
	if len(results) != len(want) {
		t.Fatalf("RunTargets len = %d, want %d", len(results), len(want))
	}
	for index, label := range want {
		if !results[index].Ran {
			t.Fatalf("results[%d].Ran = false, want true", index)
		}
		if results[index].Value != label {
			t.Fatalf("results[%d].Value = %q, want %q", index, results[index].Value, label)
		}
	}
}

func TestWorkspaceHelperProcess(t *testing.T) {
	if len(os.Args) < 4 || os.Args[1] != "-test.run=TestWorkspaceHelperProcess" {
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
