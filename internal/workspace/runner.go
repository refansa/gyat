package workspace

import (
	"fmt"
	"os/exec"
	"strings"
)

// Command describes an external command to run across workspace targets.
type Command struct {
	Name string
	Args []string
}

// Display formats the command for logs and errors.
func (command Command) Display() string {
	parts := append([]string{command.Name}, command.Args...)
	return strings.Join(parts, " ")
}

// RunOptions controls fan-out execution behavior.
type RunOptions struct {
	ContinueOnError bool
}

// RunResult is the captured output for one workspace target.
type RunResult struct {
	Target Target
	Output string
	Err    error
}

// RunCommand executes command in each target directory in order.
func RunCommand(targets []Target, command Command, options RunOptions) ([]RunResult, error) {
	if command.Name == "" {
		return nil, fmt.Errorf("command name is required")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets to execute")
	}

	results := make([]RunResult, 0, len(targets))
	failures := 0

	for _, target := range targets {
		cmd := exec.Command(command.Name, command.Args...)
		cmd.Dir = target.Dir
		out, err := cmd.CombinedOutput()

		result := RunResult{
			Target: target,
			Output: strings.TrimRight(string(out), "\r\n"),
			Err:    err,
		}
		results = append(results, result)

		if err == nil {
			continue
		}

		failures++
		if !options.ContinueOnError {
			return results, fmt.Errorf("running %q in %q: %w", command.Display(), target.Label, err)
		}
	}

	if failures > 0 {
		return results, fmt.Errorf("command failed in %d target(s)", failures)
	}

	return results, nil
}
