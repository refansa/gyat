package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <path>",
	Aliases: []string{"rm"},
	Short:   "Remove a submodule cleanly",
	Long: `Remove a submodule from the repository.

This performs the full three-step removal that git requires:
  1. Deinitialize the submodule (git submodule deinit)
  2. Delete the cached module data (.git/modules/<path>)
  3. Remove the submodule from the index (git rm)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		return runRemove(dir, cmd, args)
	},
}

func runRemove(dir string, cmd *cobra.Command, args []string) error {
	path := filepath.Clean(args[0])

	// Step 1: deinit the submodule
	fmt.Fprintf(cmd.ErrOrStderr(), "deinitializing submodule '%s'...\n", path)
	if _, err := git.Run(dir, "submodule", "deinit", "-f", path); err != nil {
		return fmt.Errorf("failed to deinit submodule: %w", err)
	}

	// Step 2: remove the cached module data from .git/modules
	modulesPath := filepath.Join(dir, ".git", "modules", path)
	if err := os.RemoveAll(modulesPath); err != nil {
		return fmt.Errorf("failed to remove module cache: %w", err)
	}

	// Step 3: remove the submodule from the index and working tree
	if _, err := git.Run(dir, "rm", "-f", path); err != nil {
		return fmt.Errorf("failed to remove submodule from index: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "removed submodule '%s'\n", path)
	fmt.Fprintln(cmd.ErrOrStderr(), "hint: don't forget to commit the changes to .gitmodules and the index")
	return nil
}
