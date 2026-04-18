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

type rmTargetResult struct {
	message string
}

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
		return runRm(startDir, dir, sharedTargetFlags, rmCached, rmForce, rmRecursive, cmd, args)
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
// runRm is the primary implementation that accepts an explicit start
// directory and explicit workspace flags.
func runRm(startDir, dir string, flags workspaceTargetFlags, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}
	if flags.hasSelection() {
		return removeSelectedWorkspaceTargets(ws, flags, cached, force, recursive, cmd, args)
	}
	return runRmWorkspace(ws, startDir, flags, cached, force, recursive, cmd, args)
}

func runRmWithoutFlags(startDir, dir string, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	return runRm(startDir, dir, workspaceTargetFlags{}, cached, force, recursive, cmd, args)
}

func runRmWorkspace(ws workspace.Workspace, startDir string, flags workspaceTargetFlags, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	rootArgs, repoTargets, err := classifyWorkspaceArgs(ws.RootDir, ws.Manifest.Repos, startDir, args)
	if err != nil {
		return err
	}
	var failures commandFailures

	if len(rootArgs) > 0 {
		gitArgs := buildRmArgs(cached, force, recursive, rootArgs)
		if _, err := git.Run(ws.RootDir, gitArgs...); err != nil {
			if handledErr := failures.handle(flags.continueOnError, "removing root paths: %w", err); handledErr != nil {
				return handledErr
			}
		}
	}

	var repos []string
	for repoPath := range repoTargets {
		repos = append(repos, repoPath)
	}
	sort.Strings(repos)

	targets := make([]workspace.Target, 0, len(repos))
	for _, repoPath := range repos {
		repo, ok := workspaceRepoByPath(ws, repoPath)
		if !ok {
			return fmt.Errorf("tracked repository '%s' not found in manifest", repoPath)
		}
		targets = append(targets, workspace.Target{
			Label:  repo.Path,
			Dir:    filepath.Join(ws.RootDir, filepath.FromSlash(repo.Path)),
			Name:   repo.Name,
			Path:   repo.Path,
			Groups: append([]string(nil), repo.Groups...),
		})
	}
	if len(targets) == 0 {
		return failures.err("rm failed")
	}

	results, err := workspace.RunTargets(targets, flags.runOptions(), func(target workspace.Target) (rmTargetResult, error) {
		stage := repoTargets[target.Path]
		if stage.stageAll {
			return rmTargetResult{}, fmt.Errorf("to remove tracked repository '%s', use 'gyat untrack %s'", target.Path, target.Path)
		}

		if _, err := os.Stat(target.Dir); os.IsNotExist(err) {
			return rmTargetResult{message: fmt.Sprintf("warning: tracked repository '%s' is not cloned, skipping\n", target.Path)}, nil
		} else if err != nil {
			return rmTargetResult{}, fmt.Errorf("checking target '%s': %w", target.Label, err)
		}

		result := rmTargetResult{message: fmt.Sprintf("removing in '%s'...\n", target.Path)}
		gitArgs := buildRmArgs(cached, force, recursive, stage.files)
		if _, err := git.Run(target.Dir, gitArgs...); err != nil {
			return result, fmt.Errorf("removing in '%s': %w", target.Path, err)
		}
		return result, nil
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
		}
	}

	return failures.err("rm failed")
}

func removeSelectedWorkspaceTargets(ws workspace.Workspace, flags workspaceTargetFlags, cached, force, recursive bool, cmd *cobra.Command, args []string) error {
	targets, err := ws.ResolveTargets(flags.targetOptions(true, nil))
	if err != nil {
		return err
	}
	var failures commandFailures
	results, err := workspace.RunTargets(targets, flags.runOptions(), func(target workspace.Target) (rmTargetResult, error) {
		label := target.Label
		if target.IsRoot {
			label = "umbrella repository"
		} else if _, err := os.Stat(target.Dir); os.IsNotExist(err) {
			return rmTargetResult{message: fmt.Sprintf("warning: tracked repository '%s' is not cloned, skipping\n", target.Path)}, nil
		} else if err != nil {
			return rmTargetResult{}, fmt.Errorf("checking target '%s': %w", target.Label, err)
		}

		result := rmTargetResult{message: fmt.Sprintf("removing in '%s'...\n", label)}
		gitArgs := buildRmArgs(cached, force, recursive, args)
		if _, err := git.Run(target.Dir, gitArgs...); err != nil {
			return result, fmt.Errorf("removing in '%s': %w", label, err)
		}
		return result, nil
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
		}
	}

	return failures.err("rm failed")
}

func init() {
	bindWorkspaceTargetFlags(rmCmd)
	bindWorkspaceParallelFlag(rmCmd)
	rmCmd.Flags().BoolVar(&rmCached, "cached", false, "Use this option to unstage and remove paths only from the index")
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Override the up-to-date check")
	rmCmd.Flags().BoolVarP(&rmRecursive, "r", "r", false, "Allow recursive removal when a leading directory name is given")
}
