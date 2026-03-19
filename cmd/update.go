package cmd

import (
	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [path]",
	Short: "Update submodules to their latest remote commit",
	Long: `Update one or all submodules to the latest commit on their tracked remote branch.

If a path is provided, only that submodule is updated.
If no path is provided, all submodules are updated.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runUpdate(dir, cmd, args)
	},
}

func runUpdate(dir string, cmd *cobra.Command, args []string) error {
	gitArgs := []string{"-c", "protocol.file.allow=always", "submodule", "update", "--remote", "--merge"}

	if len(args) == 1 {
		gitArgs = append(gitArgs, args[0])
	}

	return git.RunInteractive(dir, gitArgs...)
}
