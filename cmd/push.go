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
		return runPushFrom(startDir, dir, pushForce, cmd, args)
	},
}

// runPush pushes local commits for the selected tracked repos and, if a
// remote is configured, the umbrella repository itself.
func runPush(dir string, force bool, cmd *cobra.Command, args []string) error {
	return runPushFrom(dir, dir, force, cmd, args)
}

func runPushFrom(startDir, dir string, force bool, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}
	return runPushWorkspace(ws, startDir, force, cmd, args)
}

func runPushWorkspace(ws workspace.Workspace, startDir string, force bool, cmd *cobra.Command, args []string) error {
	selectors, err := resolveWorkspaceRepoSelectors(ws, startDir, args)
	if err != nil {
		return err
	}

	g := []string{"push"}
	if force {
		g = append(g, "--force")
	}

	pushed := 0
	var targets []workspace.Target
	if len(ws.Manifest.Repos) > 0 {
		targets, err = ws.ResolveTargets(workspace.TargetOptions{RepoSelectors: selectors})
		if err != nil {
			return err
		}
	}

	for _, target := range targets {
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
			return fmt.Errorf("pushing '%s': %w", target.Path, err)
		}
		pushed++
	}

	if hasRemote(ws.RootDir) {
		fmt.Fprintln(cmd.ErrOrStderr(), "pushing umbrella repository...")
		if err := git.RunInteractive(ws.RootDir, g...); err != nil {
			return fmt.Errorf("pushing umbrella repository: %w", err)
		}
		pushed++
	}

	if pushed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to push")
	}

	return nil
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
