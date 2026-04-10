package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/internal/git"
	"github.com/refansa/gyat/internal/manifest"
	"github.com/refansa/gyat/internal/workspace"
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
	if err == nil {
		return runUntrackWorkspace(ws, cmd, args)
	}
	if !errors.Is(err, workspace.ErrNotFound) {
		return err
	}

	return runUntrackLegacy(dir, cmd, args)
}

func runUntrackWorkspace(ws workspace.Workspace, cmd *cobra.Command, args []string) error {
	targets, err := ws.ResolveTargets(workspace.TargetOptions{
		RepoSelectors: []string{args[0]},
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

func runUntrackLegacy(dir string, cmd *cobra.Command, args []string) error {
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
