package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var pullRebase bool

var pullCmd = &cobra.Command{
	Use:   "pull [path...]",
	Short: "Pull latest commits in submodules and the umbrella repository",
	Long: `Pull the latest commits from the remote for all or specified submodules and
the umbrella repository.

With no arguments, every checked-out submodule is pulled, then the umbrella
repository is pulled if an upstream tracking branch is configured.

With one or more path arguments only the specified submodules are pulled,
then the umbrella repository.

Each repository must be on a local branch with an upstream tracking branch.
Submodules in detached HEAD state will fail — use 'gyat update' instead to
fetch the latest remote commit for detached submodule pointers.`,
	Example: `  # Pull all submodules and the umbrella
  gyat pull

  # Pull with rebase instead of merge
  gyat pull --rebase

  # Pull specific submodules only
  gyat pull services/auth services/billing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runPull(dir, pullRebase, cmd, args)
	},
}

// runPull pulls the latest commits for the resolved target submodules and,
// if an upstream is configured, the umbrella repository itself.
func runPull(dir string, rebase bool, cmd *cobra.Command, args []string) error {
	submodulePaths, err := allSubmodulePaths(dir)
	if err != nil {
		return err
	}

	gitArgs := []string{"pull"}
	if rebase {
		gitArgs = append(gitArgs, "--rebase")
	}

	targets, err := resolveTargetPaths(submodulePaths, args)
	if err != nil {
		return err
	}

	pulled := 0

	for _, path := range targets {
		subDir := filepath.Join(dir, path)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", path)
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "pulling '%s'...\n", path)
		if err := git.RunInteractive(subDir, gitArgs...); err != nil {
			return fmt.Errorf("pulling '%s': %w", path, err)
		}
		pulled++
	}

	if hasUpstream(dir) {
		fmt.Fprintln(cmd.ErrOrStderr(), "pulling umbrella repository...")
		if err := git.RunInteractive(dir, gitArgs...); err != nil {
			return fmt.Errorf("pulling umbrella repository: %w", err)
		}
		pulled++
	}

	if pulled == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to pull")
	}

	return nil
}

// hasUpstream reports whether the current branch in dir has an upstream
// tracking branch configured.
func hasUpstream(dir string) bool {
	_, err := git.Run(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	return err == nil
}

// resolveTargetPaths returns the submodule paths to operate on. When args is
// empty every registered submodule path is returned unchanged. Otherwise each
// arg is validated against the registered paths and an error is returned for
// any arg that is not a registered submodule. Paths are normalised with
// filepath.ToSlash(filepath.Clean(p)) so that callers on Windows can pass
// either forward- or back-slash separated paths.
func resolveTargetPaths(submodulePaths, args []string) ([]string, error) {
	if len(args) == 0 {
		return submodulePaths, nil
	}

	registered := make(map[string]struct{}, len(submodulePaths))
	for _, p := range submodulePaths {
		registered[filepath.ToSlash(filepath.Clean(p))] = struct{}{}
	}

	var targets []string
	for _, arg := range args {
		norm := filepath.ToSlash(filepath.Clean(arg))
		if _, ok := registered[norm]; !ok {
			return nil, fmt.Errorf("'%s' is not a registered submodule", arg)
		}
		targets = append(targets, norm)
	}

	return targets, nil
}

func init() {
	pullCmd.Flags().BoolVarP(&pullRebase, "rebase", "r", false, "Rebase instead of merge when pulling")
}
