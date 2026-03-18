package cmd

import (
	"fmt"
	"os"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync submodule URLs from .gitmodules",
	Long: `Sync synchronizes each submodule's remote URL configuration from .gitmodules.

This is useful when a submodule's URL has changed (e.g. a repo was moved or
renamed) and you need all local clones to point to the new location.

After syncing URLs, it will also re-initialize and update any submodules that
were not yet cloned.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		return runSync(dir, cmd, args)
	},
}

func runSync(dir string, cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.ErrOrStderr(), "syncing submodule URLs...")
	if err := git.RunInteractive(dir, "submodule", "sync", "--recursive"); err != nil {
		return fmt.Errorf("failed to sync submodule URLs: %w", err)
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "updating submodules...")
	if err := git.RunInteractive(dir, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive"); err != nil {
		return fmt.Errorf("failed to update submodules after sync: %w", err)
	}

	return nil
}
