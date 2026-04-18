package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

type syncTargetResult struct {
	message string
	changed bool
}

func init() {
	bindWorkspaceTargetFlags(syncCmd)
	bindWorkspaceParallelFlag(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync [path...]",
	Short: "Sync tracked repo remotes from .gyat",
	Long: `Sync synchronizes each tracked repo's remote URL configuration from .gyat.

This is useful when a tracked repo's URL has changed (e.g. a repo was moved or
renamed) and you need local clones to point to the new location.

	It also clones any tracked repos that are missing from disk and reconciles the
	gyat-managed .gitignore block.`,
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
		return runSync(startDir, dir, sharedTargetFlags, cmd, args)
	},
}

// runSync is the primary implementation that accepts an explicit start
// directory and explicit workspace flags.
func runSync(startDir, dir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}

	repoSelectors, err := resolveWorkspaceRepoSelectors(ws, startDir, args)
	if err != nil {
		return err
	}

	targets, err := ws.ResolveTargets(flags.targetOptions(true, repoSelectors))
	if err != nil {
		return err
	}

	changed := 0
	var failures commandFailures
	repoTargets := make([]workspace.Target, 0, len(targets))

	for _, target := range targets {
		if target.IsRoot {
			updated, err := workspace.SyncGitIgnore(ws.RootDir, ws.Manifest)
			if err != nil {
				if handledErr := failures.handle(flags.continueOnError, "syncing workspace .gitignore: %w", err); handledErr != nil {
					return handledErr
				}
				continue
			}
			if updated {
				fmt.Fprintln(cmd.ErrOrStderr(), "updated .gitignore managed block")
				changed++
			}
			continue
		}

		repoTargets = append(repoTargets, target)
	}
	if len(repoTargets) == 0 {
		if changed == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "nothing to sync")
		}
		return failures.err("sync failed")
	}

	results, err := workspace.RunTargets(repoTargets, flags.runOptions(), func(target workspace.Target) (syncTargetResult, error) {
		repo, ok := workspaceRepoByPath(ws, target.Path)
		if !ok {
			return syncTargetResult{}, fmt.Errorf("tracked repository '%s' not found in manifest", target.Path)
		}

		desiredURL, err := resolveManifestRepoURL(ws.RootDir, repo.URL)
		if err != nil {
			return syncTargetResult{}, fmt.Errorf("resolving URL for '%s': %w", repo.Path, err)
		}

		if _, err := os.Stat(target.Dir); os.IsNotExist(err) {
			result := syncTargetResult{message: fmt.Sprintf("cloning '%s'...\n", repo.Path)}
			cloneArgs := []string{"clone"}
			if repo.Branch != "" {
				cloneArgs = append(cloneArgs, "--branch", repo.Branch, "--single-branch")
			}
			cloneArgs = append(cloneArgs, desiredURL, target.Dir)
			if _, err := git.Run(ws.RootDir, cloneArgs...); err != nil {
				return result, fmt.Errorf("cloning '%s': %w", repo.Path, err)
			}
			result.changed = true
			return result, nil
		} else if err != nil {
			return syncTargetResult{}, fmt.Errorf("checking target '%s': %w", target.Label, err)
		}

		currentURL, err := git.Run(target.Dir, "config", "--get", "remote.origin.url")
		currentURL = strings.TrimSpace(currentURL)
		switch {
		case err != nil || currentURL == "":
			result := syncTargetResult{message: fmt.Sprintf("configuring origin for '%s'...\n", repo.Path)}
			if _, err := git.Run(target.Dir, "remote", "add", "origin", desiredURL); err != nil {
				return result, fmt.Errorf("configuring origin for '%s': %w", repo.Path, err)
			}
			result.changed = true
			return result, nil
		case currentURL != desiredURL:
			result := syncTargetResult{message: fmt.Sprintf("syncing remote for '%s'...\n", repo.Path)}
			if _, err := git.Run(target.Dir, "remote", "set-url", "origin", desiredURL); err != nil {
				return result, fmt.Errorf("syncing remote for '%s': %w", repo.Path, err)
			}
			result.changed = true
			return result, nil
		default:
			return syncTargetResult{}, nil
		}
	})
	if err != nil {
		return err
	}

	for _, result := range results {
		if !result.Ran {
			continue
		}
		if result.Value.message != "" {
			fmt.Fprint(cmd.ErrOrStderr(), result.Value.message)
		}
		if result.Err != nil {
			if handledErr := failures.handleErr(flags.continueOnError, result.Err); handledErr != nil {
				return handledErr
			}
			continue
		}
		if result.Value.changed {
			changed++
		}
	}

	if changed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to sync")
	}

	return failures.err("sync failed")
}

func runSyncWithoutFlags(startDir, dir string, cmd *cobra.Command, args []string) error {
	return runSync(startDir, dir, workspaceTargetFlags{}, cmd, args)
}

func resolveManifestRepoURL(root, repoURL string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" || !isLocalPath(repoURL) {
		return repoURL, nil
	}
	if filepath.IsAbs(repoURL) {
		return filepath.Clean(repoURL), nil
	}
	resolved := filepath.Join(root, filepath.FromSlash(repoURL))
	return filepath.Clean(resolved), nil
}
