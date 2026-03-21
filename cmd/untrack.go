package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var untrackCmd = &cobra.Command{
	Use:   "untrack <path>",
	Short: "Untrack a submodule",
	Long: `Untrack a submodule from the repository.

This performs the full three-step cleanup that git requires to fully remove a
submodule from the working tree and repository metadata:
  1. Deinitialize the submodule (git submodule deinit)
  2. Delete the cached module data (.git/modules/<path>)
  3. Remove the submodule from the index (git rm)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runUntrack(dir, cmd, args)
	},
}

func runUntrack(dir string, cmd *cobra.Command, args []string) error {
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

	fmt.Fprintf(cmd.ErrOrStderr(), "untracked submodule '%s'\n", path)
	fmt.Fprintln(cmd.ErrOrStderr(), "hint: don't forget to commit the changes to .gitmodules and the index")
	return nil
}
