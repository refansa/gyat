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
		return runSyncWithFlagsFrom(startDir, dir, sharedTargetFlags, cmd, args)
	},
}

func runSync(dir string, cmd *cobra.Command, args []string) error {
	return runSyncWithFlagsFrom(dir, dir, workspaceTargetFlags{}, cmd, args)
}

func runSyncFrom(startDir, dir string, cmd *cobra.Command, args []string) error {
	return runSyncWithFlagsFrom(startDir, dir, workspaceTargetFlags{}, cmd, args)
}

func runSyncWithFlagsFrom(startDir, dir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
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

		repo, ok := workspaceRepoByPath(ws, target.Path)
		if !ok {
			return fmt.Errorf("tracked repository '%s' not found in manifest", target.Path)
		}

		desiredURL, err := resolveManifestRepoURL(ws.RootDir, repo.URL)
		if err != nil {
			if handledErr := failures.handle(flags.continueOnError, "resolving URL for '%s': %w", repo.Path, err); handledErr != nil {
				return handledErr
			}
			continue
		}

		if _, err := os.Stat(target.Dir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "cloning '%s'...\n", repo.Path)
			cloneArgs := []string{"clone"}
			if repo.Branch != "" {
				cloneArgs = append(cloneArgs, "--branch", repo.Branch, "--single-branch")
			}
			cloneArgs = append(cloneArgs, desiredURL, target.Dir)
			if _, err := git.Run(ws.RootDir, cloneArgs...); err != nil {
				if handledErr := failures.handle(flags.continueOnError, "cloning '%s': %w", repo.Path, err); handledErr != nil {
					return handledErr
				}
				continue
			}
			changed++
			continue
		} else if err != nil {
			return err
		}

		currentURL, err := git.Run(target.Dir, "config", "--get", "remote.origin.url")
		currentURL = strings.TrimSpace(currentURL)
		switch {
		case err != nil || currentURL == "":
			fmt.Fprintf(cmd.ErrOrStderr(), "configuring origin for '%s'...\n", repo.Path)
			if _, err := git.Run(target.Dir, "remote", "add", "origin", desiredURL); err != nil {
				if handledErr := failures.handle(flags.continueOnError, "configuring origin for '%s': %w", repo.Path, err); handledErr != nil {
					return handledErr
				}
				continue
			}
			changed++
		case currentURL != desiredURL:
			fmt.Fprintf(cmd.ErrOrStderr(), "syncing remote for '%s'...\n", repo.Path)
			if _, err := git.Run(target.Dir, "remote", "set-url", "origin", desiredURL); err != nil {
				if handledErr := failures.handle(flags.continueOnError, "syncing remote for '%s': %w", repo.Path, err); handledErr != nil {
					return handledErr
				}
				continue
			}
			changed++
		}
	}

	if changed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to sync")
	}

	return failures.err("sync failed")
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
