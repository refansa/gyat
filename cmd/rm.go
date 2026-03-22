package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var (
	rmCached    bool
	rmForce     bool
	rmRecursive bool
)

var rmCmd = &cobra.Command{
	Use:   "rm [path...]",
	Short: "Remove files from the working tree and from the index",
	Long: `Remove files from the working tree and from the index across the umbrella
repository and any registered submodules.

Behaves like 'git rm' but is submodule-aware: paths that live inside a
submodule are routed to that submodule, while all other paths are
removed from the umbrella repository itself.`,
	Example: `  # Remove a file from the umbrella repository
  gyat rm .gitignore

  # Remove a file inside a submodule
  gyat rm services/auth/handler.go

  # Remove files recursively from a submodule
  gyat rm -r services/auth/models

  # Remove files from the index only
  gyat rm --cached services/auth/README.md`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runRm(dir, rmCached, rmForce, rmRecursive, cmd, args)
	},
}

// buildRmArgs constructs the argument slice for a git rm invocation.
func buildRmArgs(cached, force, recursive bool, files []string) []string {
	args := []string{"rm"}
	if cached {
		args = append(args, "--cached")
	}
	if force {
		args = append(args, "-f")
	}
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, "--")
	args = append(args, files...)
	return args
}

// runRm removes files from the working tree and index. It routes each path to
// the repository it belongs to.
func runRm(dir string, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	submodulePaths, err := allSubmodulePaths(dir)
	if err != nil {
		return err
	}

	rootArgs, subTargets := classifyArgs(submodulePaths, args)

	if len(rootArgs) > 0 {
		gitArgs := buildRmArgs(cached, force, recursive, rootArgs)
		if _, err := git.Run(dir, gitArgs...); err != nil {
			return fmt.Errorf("removing root paths: %w", err)
		}
	}

	// Sort submodules for deterministic output (useful for tests and UX).
	var subs []string
	for sub := range subTargets {
		subs = append(subs, sub)
	}
	sort.Strings(subs)

	for _, sub := range subs {
		stage := subTargets[sub]
		subDir := filepath.Join(dir, sub)

		if stage.stageAll {
			return fmt.Errorf("to remove submodule '%s', use 'gyat untrack %s'", sub, sub)
		}

		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", sub)
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "removing in '%s'...\n", sub)

		gitArgs := buildRmArgs(cached, force, recursive, stage.files)
		if _, err := git.Run(subDir, gitArgs...); err != nil {
			return fmt.Errorf("removing in '%s': %w", sub, err)
		}
	}

	return nil
}

func init() {
	rmCmd.Flags().BoolVar(&rmCached, "cached", false, "Use this option to unstage and remove paths only from the index")
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Override the up-to-date check")
	rmCmd.Flags().BoolVarP(&rmRecursive, "r", "r", false, "Allow recursive removal when a leading directory name is given")
}
