package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [submodule-path...]",
	Short: "Stage changes in submodules",
	Long: `Stage changes in one or all submodules.

If no submodule paths are provided, all changes in every submodule are staged
(equivalent to running 'git add -A' inside each one).

If one or more submodule paths are provided, only those submodules are staged.`,
	Example: `  # Stage all changes across every submodule
  gyat add

  # Stage changes in a specific submodule
  gyat add services/auth

  # Stage changes in multiple specific submodules
  gyat add services/auth services/billing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		return runAdd(dir, cmd, args)
	},
}

// runAdd stages all working-tree changes inside each target submodule.
// When args is empty every registered submodule is targeted; otherwise only
// the submodules whose paths are listed in args are staged.
func runAdd(dir string, cmd *cobra.Command, args []string) error {
	targets, err := resolveSubmodulePaths(dir, args)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "no submodules found")
		fmt.Fprintln(cmd.ErrOrStderr(), "hint: use 'gyat track <repo>' to add a repository")
		return nil
	}

	staged := 0
	for _, path := range targets {
		subDir := filepath.Join(dir, path)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", path)
			continue
		}

		// Skip submodules with nothing to stage.
		statusOut, err := git.Run(subDir, "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("checking status of '%s': %w", path, err)
		}
		if strings.TrimSpace(statusOut) == "" {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "staging '%s'...\n", path)
		if _, err := git.Run(subDir, "add", "-A"); err != nil {
			return fmt.Errorf("staging '%s': %w", path, err)
		}
		staged++
	}

	if staged == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to stage")
	}

	return nil
}

// resolveSubmodulePaths returns the submodule paths to operate on.
// When args is non-empty it is returned as-is; otherwise all paths registered
// in .gitmodules are returned.
func resolveSubmodulePaths(dir string, args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}
	return allSubmodulePaths(dir)
}

// allSubmodulePaths reads every submodule path from .gitmodules.
// Returns nil (not an error) when .gitmodules is absent or empty.
func allSubmodulePaths(dir string) ([]string, error) {
	pathsOut, err := git.Run(dir, "config", "-f", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil || strings.TrimSpace(pathsOut) == "" {
		return nil, nil
	}

	var paths []string
	for _, line := range strings.Split(pathsOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		paths = append(paths, parts[1])
	}
	return paths, nil
}
