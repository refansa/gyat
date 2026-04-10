package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecute_DoesNotPrintUsageOnError(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	oldOut := rootCmd.OutOrStdout()
	oldErr := rootCmd.ErrOrStderr()

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"definitely-not-a-command"})
	defer func() {
		rootCmd.SetOut(oldOut)
		rootCmd.SetErr(oldErr)
		rootCmd.SetArgs(nil)
	}()

	err := Execute()
	if err == nil {
		t.Fatal("expected unknown command error")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stdout.String(), "Usage:") || strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected usage output to be suppressed\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
}
