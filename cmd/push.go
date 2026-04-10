package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var pushForce bool

var pushCmd = &cobra.Command{
	Use:   "push [path...]",
	Short: "Push commits in tracked repos and the umbrella repository",
	Long: `Push local commits to the remote for all or specified tracked repos and the
umbrella repository.

With no arguments, every cloned tracked repo is pushed, then the umbrella
repository is pushed if a remote is configured.

With one or more path arguments only the specified tracked repos are pushed,
then the umbrella repository.`,
	Example: `  # Push all tracked repos and the umbrella
  gyat push

	# Push specific tracked repos only
  gyat push services/auth services/billing

  # Force push (use with care — rewrites remote history)
  gyat push --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runPushWithFlagsFrom(startDir, dir, sharedTargetFlags, pushForce, cmd, args)
	},
}

// runPush pushes local commits for the selected tracked repos and, if a
// remote is configured, the umbrella repository itself.
func runPush(dir string, force bool, cmd *cobra.Command, args []string) error {
	return runPushWithFlagsFrom(dir, dir, workspaceTargetFlags{}, force, cmd, args)
}

func runPushFrom(startDir, dir string, force bool, cmd *cobra.Command, args []string) error {
	return runPushWithFlagsFrom(startDir, dir, workspaceTargetFlags{}, force, cmd, args)
}

func runPushWithFlagsFrom(startDir, dir string, flags workspaceTargetFlags, force bool, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}
	return runPushWorkspace(ws, startDir, flags, force, cmd, args)
}

func runPushWorkspace(ws workspace.Workspace, startDir string, flags workspaceTargetFlags, force bool, cmd *cobra.Command, args []string) error {
	repoSelectors, err := resolveWorkspaceRepoSelectors(ws, startDir, args)
	if err != nil {
		return err
	}

	g := []string{"push"}
	if force {
		g = append(g, "--force")
	}

	targets, err := ws.ResolveTargets(flags.targetOptions(true, repoSelectors))
	if err != nil {
		return err
	}

	pushed := 0
	var failures commandFailures
	for _, target := range targets {
		if target.IsRoot {
			if !hasRemote(ws.RootDir) {
				continue
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "pushing umbrella repository...")
			if err := git.RunInteractive(ws.RootDir, g...); err != nil {
				if handledErr := failures.handle(flags.continueOnError, "pushing umbrella repository: %w", err); handledErr != nil {
					return handledErr
				}
				continue
			}
			pushed++
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

		fmt.Fprintf(cmd.ErrOrStderr(), "pushing '%s'...\n", target.Path)
		if err := git.RunInteractive(target.Dir, g...); err != nil {
			if handledErr := failures.handle(flags.continueOnError, "pushing '%s': %w", target.Path, err); handledErr != nil {
				return handledErr
			}
			continue
		}
		pushed++
	}

	if pushed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to push")
	}

	return failures.err("push failed")
}

// hasRemote reports whether the git repository in dir has at least one remote
// configured.
func hasRemote(dir string) bool {
	out, err := git.Run(dir, "remote")
	return err == nil && strings.TrimSpace(out) != ""
}

func init() {
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "Force push (use with care — rewrites remote history)")
}
