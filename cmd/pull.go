package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var pullRebase bool

var pullCmd = &cobra.Command{
	Use:   "pull [path...]",
	Short: "Pull latest commits in tracked repos and the umbrella repository",
	Long: `Pull the latest commits from the remote for all or specified tracked repos and
the umbrella repository.

The command walks the selected targets in deterministic order.

With no arguments, every cloned tracked repo is pulled, then the umbrella
repository is pulled if an upstream tracking branch is configured.

With one or more path arguments only the specified tracked repos are pulled,
then the umbrella repository.

Each repository must be on a local branch with an upstream tracking branch.
Tracked repos in detached HEAD state will fail — use 'gyat update' instead to
fetch the latest remote commit for a detached repository.`,
	Example: `  # Pull all tracked repos and the umbrella
  gyat pull

  # Pull with rebase instead of merge
  gyat pull --rebase

	# Pull specific tracked repos only
  gyat pull services/auth services/billing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runPullWithFlagsFrom(startDir, dir, sharedTargetFlags, pullRebase, cmd, args)
	},
}

// runPull pulls the latest commits for the selected tracked repos and, if an
// upstream is configured, the umbrella repository itself.
func runPull(dir string, rebase bool, cmd *cobra.Command, args []string) error {
	return runPullWithFlagsFrom(dir, dir, workspaceTargetFlags{}, rebase, cmd, args)
}

func runPullFrom(startDir, dir string, rebase bool, cmd *cobra.Command, args []string) error {
	return runPullWithFlagsFrom(startDir, dir, workspaceTargetFlags{}, rebase, cmd, args)
}

func runPullWithFlagsFrom(startDir, dir string, flags workspaceTargetFlags, rebase bool, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}
	return runPullWorkspace(ws, startDir, flags, rebase, cmd, args)
}

func runPullWorkspace(ws workspace.Workspace, startDir string, flags workspaceTargetFlags, rebase bool, cmd *cobra.Command, args []string) error {
	repoSelectors, err := resolveWorkspaceRepoSelectors(ws, startDir, args)
	if err != nil {
		return err
	}

	g := []string{"pull"}
	if rebase {
		g = append(g, "--rebase")
	}

	targets, err := ws.ResolveTargets(flags.targetOptions(true, repoSelectors))
	if err != nil {
		return err
	}

	pulled := 0
	var failures commandFailures
	for _, target := range targets {
		if target.IsRoot {
			if !hasUpstream(ws.RootDir) {
				continue
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "pulling umbrella repository...")
			if err := git.RunInteractive(ws.RootDir, g...); err != nil {
				if handledErr := failures.handle(flags.continueOnError, "pulling umbrella repository: %w", err); handledErr != nil {
					return handledErr
				}
				continue
			}
			pulled++
			continue
		}

		repo, ok := workspaceRepoByPath(ws, target.Path)
		if !ok {
			return fmt.Errorf("tracked repository '%s' not found in manifest", target.Path)
		}
		if isLocalPath(repo.URL) {
			fmt.Fprintf(cmd.ErrOrStderr(), "hint: '%s' uses a local path remote — skipping\n", target.Path)
			continue
		}
		if _, err := os.Stat(target.Dir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", target.Path)
			continue
		} else if err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "pulling '%s'...\n", target.Path)
		if err := git.RunInteractive(target.Dir, g...); err != nil {
			if handledErr := failures.handle(flags.continueOnError, "pulling '%s': %w", target.Path, err); handledErr != nil {
				return handledErr
			}
			continue
		}
		pulled++
	}

	if pulled == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to pull")
	}

	return failures.err("pull failed")
}

func resolveWorkspaceRepoSelectors(ws workspace.Workspace, startDir string, args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}

	selectors := make([]string, 0, len(args))
	seen := make(map[string]struct{}, len(args))

	for _, arg := range args {
		repoPath, err := resolveWorkspaceRepoSelector(ws, startDir, arg)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[repoPath]; exists {
			continue
		}
		seen[repoPath] = struct{}{}
		selectors = append(selectors, repoPath)
	}

	return selectors, nil
}

func resolveWorkspaceRepoSelector(ws workspace.Workspace, startDir, arg string) (string, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return "", fmt.Errorf("path is required")
	}

	for _, repo := range ws.Manifest.Repos {
		if repo.Name == trimmed {
			return repo.Path, nil
		}
	}

	rel, err := normalizeWorkspaceArg(ws.RootDir, startDir, trimmed)
	if err != nil {
		return "", err
	}

	repoPath, _, _, matched := matchTrackedRepo(ws.Manifest.Repos, rel)
	if matched {
		return repoPath, nil
	}

	return "", fmt.Errorf("'%s' is not a tracked repository", arg)
}

func workspaceRepoByPath(ws workspace.Workspace, path string) (manifest.Repo, bool) {
	for _, repo := range ws.Manifest.Repos {
		if repo.Path == path {
			return repo, true
		}
	}
	return manifest.Repo{}, false
}

// hasUpstream reports whether the current branch in dir has an upstream
// tracking branch configured.
func hasUpstream(dir string) bool {
	_, err := git.Run(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	return err == nil
}

func init() {
	pullCmd.Flags().BoolVarP(&pullRebase, "rebase", "r", false, "Rebase instead of merge when pulling")
}
