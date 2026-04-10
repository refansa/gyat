package cmd

import (
	"fmt"
	"os"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var untrackCmd = &cobra.Command{
	Use:   "untrack <path>",
	Short: "Remove a tracked repository from the current gyat workspace",
	Long: `Remove a tracked repository from the current gyat workspace.

This removes the repository from the .gyat manifest, deletes its working-tree
directory, and updates the gyat-managed block in .gitignore.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runUntrack(dir, cmd, args)
	},
}

func runUntrack(dir string, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runUntrackWorkspace(ws, dir, cmd, args)
}

func runUntrackWorkspace(ws workspace.Workspace, startDir string, cmd *cobra.Command, args []string) error {
	repoPath, err := resolveWorkspaceRepoSelector(ws, startDir, args[0])
	if err != nil {
		return err
	}

	targets, err := ws.ResolveTargets(workspace.TargetOptions{
		RepoSelectors: []string{repoPath},
	})
	if err != nil {
		return err
	}
	if len(targets) != 1 || targets[0].IsRoot {
		return fmt.Errorf("'%s' is not a tracked repository", args[0])
	}

	target := targets[0]
	fmt.Fprintf(cmd.ErrOrStderr(), "removing tracked repository '%s'...\n", target.Path)
	if err := os.RemoveAll(target.Dir); err != nil {
		return fmt.Errorf("removing repository working tree: %w", err)
	}

	updated := ws.Manifest
	updated.Repos = removeTrackedRepo(updated.Repos, target.Path)
	if err := manifest.SaveDir(ws.RootDir, updated); err != nil {
		return err
	}
	if changed, err := workspace.SyncGitIgnore(ws.RootDir, updated); err != nil {
		return err
	} else if changed {
		fmt.Fprintln(cmd.ErrOrStderr(), "updated .gitignore managed block")
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "untracked repository '%s'\n", target.Path)
	fmt.Fprintln(cmd.ErrOrStderr(), "hint: commit the changes to .gyat and .gitignore")
	return nil
}

func removeTrackedRepo(repos []manifest.Repo, targetPath string) []manifest.Repo {
	filtered := repos[:0]
	for _, repo := range repos {
		if repo.Path == targetPath {
			continue
		}
		filtered = append(filtered, repo)
	}
	return filtered
}
