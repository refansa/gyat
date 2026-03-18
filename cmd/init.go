package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a gyat-managed repository",
	Long: `Initialize a new git repository in the current directory, or reinitialize
an existing one. If a .gitmodules file is already present (e.g. after cloning
an existing gyat-managed repo), all submodules will be initialized and checked
out automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		return runInit(dir, cmd, args)
	},
}

func runInit(dir string, cmd *cobra.Command, args []string) error {
	out, err := git.Run(dir, "init")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.ErrOrStderr(), out)

	// If .gitmodules exists, we're likely in a freshly cloned umbrella repo.
	// Initialize and check out all submodules automatically.
	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); err == nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "found .gitmodules — initializing submodules...")
		if err := git.RunInteractive(dir, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive"); err != nil {
			return fmt.Errorf("failed to initialize submodules: %w", err)
		}
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "hint: use 'gyat add <repo> [path]' to start adding repositories")
	}

	return nil
}
