package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [path...]",
	Short: "Update tracked repos to their latest remote commit",
	Long: `Update one or more tracked repos to the latest commit on their tracked remote branch.

If paths are provided, only the selected tracked repos are updated.
If no paths are provided, all tracked repos are updated.

Repos with a configured branch in .gyat are updated from origin/<branch> using
a fast-forward pull. Repos without an explicit branch use their current
tracking branch.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runUpdateFrom(startDir, dir, cmd, args)
	},
}

func runUpdate(dir string, cmd *cobra.Command, args []string) error {
	return runUpdateFrom(dir, dir, cmd, args)
}

func runUpdateFrom(startDir, dir string, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}

	selectors, err := resolveWorkspaceRepoSelectors(ws, startDir, args)
	if err != nil {
		return err
	}

	if len(ws.Manifest.Repos) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to update")
		return nil
	}

	targets, err := ws.ResolveTargets(workspace.TargetOptions{RepoSelectors: selectors})
	if err != nil {
		return err
	}

	updated := 0
	for _, target := range targets {
		repo, ok := workspaceRepoByPath(ws, target.Path)
		if !ok {
			return fmt.Errorf("tracked repository '%s' not found in manifest", target.Path)
		}

		repoDir := filepath.Join(ws.RootDir, filepath.FromSlash(repo.Path))
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", repo.Path)
			continue
		} else if err != nil {
			return err
		}

		gitArgs := []string{"-c", "protocol.file.allow=always", "pull", "--ff-only"}
		if repo.Branch != "" {
			gitArgs = append(gitArgs, "origin", repo.Branch)
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "updating '%s'...\n", repo.Path)
		if err := git.RunInteractive(repoDir, gitArgs...); err != nil {
			return fmt.Errorf("updating '%s': %w", repo.Path, err)
		}
		updated++
	}

	if updated == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to update")
	}

	return nil
}
