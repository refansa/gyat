package cmd

import (
	"fmt"
	"os"

	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var untrackCmd = &cobra.Command{
	Use:   "untrack [path...]",
	Short: "Remove a tracked repository from the current gyat workspace",
	Long: `Remove a tracked repository from the current gyat workspace.

This removes the repository from the .gyat manifest, deletes its working-tree
directory, and updates the gyat-managed block in .gitignore.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runUntrackWithFlags(dir, sharedTargetFlags, cmd, args)
	},
}

func runUntrack(dir string, cmd *cobra.Command, args []string) error {
	return runUntrackWithFlags(dir, workspaceTargetFlags{}, cmd, args)
}

func runUntrackWithFlags(dir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runUntrackWorkspace(ws, dir, flags, cmd, args)
}

func runUntrackWorkspace(ws workspace.Workspace, startDir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	if flags.rootOnly {
		return fmt.Errorf("untrack does not support --root-only")
	}

	repoSelectors, err := resolveWorkspaceRepoSelectors(ws, startDir, args)
	if err != nil {
		return err
	}
	if len(repoSelectors) == 0 && len(flags.repoSelectors) == 0 && len(flags.groups) == 0 {
		return fmt.Errorf("at least one tracked repository is required")
	}

	targets, err := ws.ResolveTargets(flags.targetOptions(false, repoSelectors))
	if err != nil {
		return err
	}

	removedPaths := make([]string, 0, len(targets))
	var failures commandFailures
	for _, target := range targets {
		if target.IsRoot {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "removing tracked repository '%s'...\n", target.Path)
		if err := os.RemoveAll(target.Dir); err != nil {
			if handledErr := failures.handle(flags.continueOnError, "removing repository '%s': %w", target.Path, err); handledErr != nil {
				return handledErr
			}
			continue
		}

		removedPaths = append(removedPaths, target.Path)
		fmt.Fprintf(cmd.ErrOrStderr(), "untracked repository '%s'\n", target.Path)
	}

	if len(removedPaths) > 0 {
		updated := ws.Manifest
		for _, removedPath := range removedPaths {
			updated.Repos = removeTrackedRepo(updated.Repos, removedPath)
		}
		if err := manifest.SaveDir(ws.RootDir, updated); err != nil {
			return err
		}
		if changed, err := workspace.SyncGitIgnore(ws.RootDir, updated); err != nil {
			return err
		} else if changed {
			fmt.Fprintln(cmd.ErrOrStderr(), "updated .gitignore managed block")
		}
		fmt.Fprintln(cmd.ErrOrStderr(), "hint: commit the changes to .gyat and .gitignore")
	}

	return failures.err("untrack failed")
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
