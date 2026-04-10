package workspace

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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
	Parallel        bool
}

// TargetResult is the captured output for one target callback execution.
type TargetResult[T any] struct {
	Target Target
	Value  T
	Err    error
	Ran    bool
}

// RunTargets executes run for each target and returns ordered results.
func RunTargets[T any](targets []Target, options RunOptions, run func(Target) (T, error)) ([]TargetResult[T], error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets to execute")
	}
	if run == nil {
		return nil, fmt.Errorf("target runner is required")
	}

	results := make([]TargetResult[T], len(targets))

	if !options.Parallel || len(targets) == 1 {
		for index, target := range targets {
			value, err := run(target)
			results[index] = TargetResult[T]{
				Target: target,
				Value:  value,
				Err:    err,
				Ran:    true,
			}
			if err != nil && !options.ContinueOnError {
				break
			}
		}
		return results, nil
	}

	workerCount := runtime.GOMAXPROCS(0)
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(targets) {
		workerCount = len(targets)
	}

	var next atomic.Int64
	var stop atomic.Bool
	var wg sync.WaitGroup

	wg.Add(workerCount)
	for range workerCount {
		go func() {
			defer wg.Done()

			for {
				if !options.ContinueOnError && stop.Load() {
					return
				}

				index := int(next.Add(1)) - 1
				if index >= len(targets) {
					return
				}

				target := targets[index]
				value, err := run(target)
				results[index] = TargetResult[T]{
					Target: target,
					Value:  value,
					Err:    err,
					Ran:    true,
				}

				if err != nil && !options.ContinueOnError {
					stop.Store(true)
				}
			}
		}()
	}

	wg.Wait()
	return results, nil
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

	runResults, err := RunTargets(targets, options, func(target Target) (string, error) {
		cmd := exec.Command(command.Name, command.Args...)
		cmd.Dir = target.Dir
		out, err := cmd.CombinedOutput()
		return strings.TrimRight(string(out), "\r\n"), err
	})
	if err != nil {
		return nil, err
	}

	results := make([]RunResult, 0, len(runResults))
	failures := 0
	for _, runResult := range runResults {
		if !runResult.Ran {
			continue
		}

		result := RunResult{
			Target: runResult.Target,
			Output: runResult.Value,
			Err:    runResult.Err,
		}
		results = append(results, result)

		if runResult.Err == nil {
			continue
		}

		failures++
		if !options.ContinueOnError {
			return results, fmt.Errorf("running %q in %q: %w", command.Display(), runResult.Target.Label, runResult.Err)
		}
	}

	if failures > 0 {
		return results, fmt.Errorf("command failed in %d target(s)", failures)
	}

	return results, nil
}
