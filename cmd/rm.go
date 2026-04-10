package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	rmCached    bool
	rmForce     bool
	rmRecursive bool
)

var rmCmd = &cobra.Command{
	Use:   "rm [path...]",
	Short: "Remove files from the working tree and from the index",
	Long: `Remove files from the working tree and from the index across the umbrella
repository and tracked repos.

Behaves like 'git rm' but is workspace-aware: paths that live inside a
tracked repo are routed to that repo, while all other paths are
removed from the umbrella repository itself.`,
	Example: `  # Remove a file from the umbrella repository
  gyat rm .gitignore

	# Remove a file inside a tracked repo
  gyat rm services/auth/handler.go

	# Remove files recursively from a tracked repo
  gyat rm -r services/auth/models

	# Remove files from the index only
  gyat rm --cached services/auth/README.md`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runRmFrom(startDir, dir, rmCached, rmForce, rmRecursive, cmd, args)
	},
}

// buildRmArgs constructs the argument slice for a git rm invocation.
func buildRmArgs(cached, force, recursive bool, files []string) []string {
	args := []string{"rm"}
	if cached {
		args = append(args, "--cached")
	}
	if force {
		args = append(args, "-f")
	}
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, "--")
	args = append(args, files...)
	return args
}

// runRm removes files from the working tree and index. It routes each path to
// the repository it belongs to.
func runRm(dir string, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	return runRmFrom(dir, dir, cached, force, recursive, cmd, args)
}

func runRmFrom(startDir, dir string, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}

	rootArgs, repoTargets, err := classifyWorkspaceArgs(ws.RootDir, ws.Manifest.Repos, startDir, args)
	if err != nil {
		return err
	}

	if len(rootArgs) > 0 {
		gitArgs := buildRmArgs(cached, force, recursive, rootArgs)
		if _, err := git.Run(ws.RootDir, gitArgs...); err != nil {
			return fmt.Errorf("removing root paths: %w", err)
		}
	}

	var repos []string
	for repoPath := range repoTargets {
		repos = append(repos, repoPath)
	}
	sort.Strings(repos)

	for _, repoPath := range repos {
		stage := repoTargets[repoPath]
		repo, ok := workspaceRepoByPath(ws, repoPath)
		if !ok {
			return fmt.Errorf("tracked repository '%s' not found in manifest", repoPath)
		}
		repoDir := filepath.Join(ws.RootDir, filepath.FromSlash(repo.Path))

		if stage.stageAll {
			return fmt.Errorf("to remove tracked repository '%s', use 'gyat untrack %s'", repoPath, repoPath)
		}

		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", repoPath)
			continue
		} else if err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "removing in '%s'...\n", repoPath)

		gitArgs := buildRmArgs(cached, force, recursive, stage.files)
		if _, err := git.Run(repoDir, gitArgs...); err != nil {
			return fmt.Errorf("removing in '%s': %w", repoPath, err)
		}
	}

	return nil
}

func init() {
	rmCmd.Flags().BoolVar(&rmCached, "cached", false, "Use this option to unstage and remove paths only from the index")
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Override the up-to-date check")
	rmCmd.Flags().BoolVarP(&rmRecursive, "r", "r", false, "Allow recursive removal when a leading directory name is given")
}
