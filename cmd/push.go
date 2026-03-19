package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var pushForce bool

var pushCmd = &cobra.Command{
	Use:   "push [path...]",
	Short: "Push commits in submodules and the umbrella repository",
	Long: `Push local commits to the remote for all or specified submodules and the
umbrella repository.

With no arguments, every checked-out submodule is pushed, then the umbrella
repository is pushed if a remote is configured.

With one or more path arguments only the specified submodules are pushed,
then the umbrella repository.`,
	Example: `  # Push all submodules and the umbrella
  gyat push

  # Push specific submodules only
  gyat push services/auth services/billing

  # Force push (use with care — rewrites remote history)
  gyat push --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runPush(dir, pushForce, cmd, args)
	},
}

// runPush pushes local commits for the resolved target submodules and,
// if a remote is configured, the umbrella repository itself.
func runPush(dir string, force bool, cmd *cobra.Command, args []string) error {
	submodulePaths, err := allSubmodulePaths(dir)
	if err != nil {
		return err
	}

	gitArgs := []string{"push"}
	if force {
		gitArgs = append(gitArgs, "--force")
	}

	targets, err := resolveTargetPaths(submodulePaths, args)
	if err != nil {
		return err
	}

	pushed := 0

	for _, path := range targets {
		subDir := filepath.Join(dir, path)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", path)
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "pushing '%s'...\n", path)
		if err := git.RunInteractive(subDir, gitArgs...); err != nil {
			return fmt.Errorf("pushing '%s': %w", path, err)
		}
		pushed++
	}

	if hasRemote(dir) {
		fmt.Fprintln(cmd.ErrOrStderr(), "pushing umbrella repository...")
		if err := git.RunInteractive(dir, gitArgs...); err != nil {
			return fmt.Errorf("pushing umbrella repository: %w", err)
		}
		pushed++
	}

	if pushed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to push")
	}

	return nil
}

// hasRemote reports whether the git repository in dir has at least one remote
// configured.
func hasRemote(dir string) bool {
	out, err := git.Run(dir, "remote")
	return err == nil && strings.TrimSpace(out) != ""
}

func init() {
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "Force push (use with care — rewrites remote history)")
}
