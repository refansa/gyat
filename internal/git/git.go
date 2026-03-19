package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a git command in dir and returns its stdout output.
// Pass an empty string for dir to use the current working directory.
// If the command fails, the error will contain the stderr message.
func Run(dir string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}

	return strings.TrimRight(stdout.String(), "\r\n"), nil
}

// RunInteractive executes a git command in dir with stdin, stdout, and stderr
// attached to the current process. Use this for commands that show progress
// bars, prompt for credentials, or otherwise need a live terminal.
// Pass an empty string for dir to use the current working directory.
func RunInteractive(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
