package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var (
	commitMessage  string
	commitNoVerify bool
)

var commitCmd = &cobra.Command{
	Use:   "commit [path...]",
	Short: "Commit staged changes across submodules and the umbrella repository",
	Long: `Commit staged changes in submodules and the umbrella repository.

With no arguments, every checked-out submodule that has staged changes is
committed, then the updated submodule refs are staged in the umbrella
repository and the umbrella itself is committed — all with the same message.

With one or more path arguments only the specified submodules are committed.
The umbrella repository is still committed afterwards if submodule refs
were updated.`,
	Example: `  # Commit all submodules with staged changes, then the umbrella
  gyat commit -m "feat: add login endpoint"

  # Commit only specific submodules
  gyat commit -m "fix: typo" services/auth services/billing

  # Skip git hooks
  gyat commit -m "wip" --no-verify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if commitMessage == "" {
			return fmt.Errorf("required flag \"message\" not set")
		}
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runCommit(dir, commitMessage, commitNoVerify, cmd, args)
	},
}

// hasStagedChanges reports whether git status --porcelain output contains any
// line representing a staged (index) change. In the two-column porcelain
// format "XY PATH", position 0 (X) is the index status. If X is not ' ' and
// not '?', there is a staged change.
func hasStagedChanges(statusOut string) bool {
	for _, line := range strings.Split(statusOut, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 2 {
			continue
		}
		x := line[0]
		if x != ' ' && x != '?' {
			return true
		}
	}
	return false
}

// buildCommitArgs constructs the argument slice for a git commit invocation.
func buildCommitArgs(message string, noVerify bool) []string {
	args := []string{"commit", "-m", message}
	if noVerify {
		args = append(args, "--no-verify")
	}
	return args
}

// runCommit commits staged changes in targeted (or all) submodules and then
// commits the umbrella repository. The function signature mirrors runTrack so
// that tests can invoke it directly without touching global flag state.
func runCommit(dir, message string, noVerify bool, cmd *cobra.Command, args []string) error {
	submodulePaths, err := allSubmodulePaths(dir)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return commitAll(dir, message, noVerify, submodulePaths, cmd)
	}
	return commitTargeted(dir, message, noVerify, submodulePaths, args, cmd)
}

// commitAll iterates over every registered submodule path, commits those with
// staged changes, stages the updated refs in the umbrella, and commits the
// umbrella.
func commitAll(dir, message string, noVerify bool, submodulePaths []string, cmd *cobra.Command) error {
	committed := 0

	for _, sub := range submodulePaths {
		subDir := filepath.Join(dir, sub)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", sub)
			continue
		}

		statusOut, err := git.Run(subDir, "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("checking status of '%s': %w", sub, err)
		}
		if !hasStagedChanges(statusOut) {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "committing '%s'...\n", sub)
		commitArgs := buildCommitArgs(message, noVerify)
		if _, err := git.Run(subDir, commitArgs...); err != nil {
			return fmt.Errorf("committing '%s': %w", sub, err)
		}
		committed++

		// Stage the updated submodule ref in the umbrella.
		if _, err := git.Run(dir, "add", sub); err != nil {
			return fmt.Errorf("staging submodule ref '%s': %w", sub, err)
		}
	}

	// Commit the umbrella if it has staged changes.
	umbrellaStatus, err := git.Run(dir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("checking umbrella status: %w", err)
	}
	if hasStagedChanges(umbrellaStatus) {
		fmt.Fprintln(cmd.ErrOrStderr(), "committing umbrella repository...")
		commitArgs := buildCommitArgs(message, noVerify)
		if _, err := git.Run(dir, commitArgs...); err != nil {
			return fmt.Errorf("committing umbrella repository: %w", err)
		}
		committed++
	}

	if committed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to commit")
	}

	return nil
}

// commitTargeted commits only the submodules specified by args, stages their
// updated refs in the umbrella, and commits the umbrella.
func commitTargeted(dir, message string, noVerify bool, submodulePaths, args []string, cmd *cobra.Command) error {
	registered := make(map[string]bool, len(submodulePaths))
	for _, p := range submodulePaths {
		registered[filepath.ToSlash(filepath.Clean(p))] = true
	}

	committed := 0

	for _, arg := range args {
		norm := filepath.ToSlash(filepath.Clean(arg))
		if !registered[norm] {
			return fmt.Errorf("'%s' is not a registered submodule", arg)
		}

		subDir := filepath.Join(dir, norm)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", norm)
			continue
		}

		statusOut, err := git.Run(subDir, "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("checking status of '%s': %w", norm, err)
		}
		if !hasStagedChanges(statusOut) {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "committing '%s'...\n", norm)
		commitArgs := buildCommitArgs(message, noVerify)
		if _, err := git.Run(subDir, commitArgs...); err != nil {
			return fmt.Errorf("committing '%s': %w", norm, err)
		}
		committed++

		// Stage the updated submodule ref in the umbrella.
		if _, err := git.Run(dir, "add", norm); err != nil {
			return fmt.Errorf("staging submodule ref '%s': %w", norm, err)
		}
	}

	// Commit the umbrella if it has staged changes after staging refs.
	umbrellaStatus, err := git.Run(dir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("checking umbrella status: %w", err)
	}
	if hasStagedChanges(umbrellaStatus) {
		fmt.Fprintln(cmd.ErrOrStderr(), "committing umbrella repository...")
		commitArgs := buildCommitArgs(message, noVerify)
		if _, err := git.Run(dir, commitArgs...); err != nil {
			return fmt.Errorf("committing umbrella repository: %w", err)
		}
		committed++
	}

	if committed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to commit")
	}

	return nil
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message (required)")
	commitCmd.Flags().BoolVar(&commitNoVerify, "no-verify", false, "Bypass git hooks")
}
