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

var (
	commitMessage  string
	commitNoVerify bool
)

var commitCmd = &cobra.Command{
	Use:   "commit [path...]",
	Short: "Commit staged changes across tracked repos and the umbrella repository",
	Long: `Commit staged changes in tracked repos and the umbrella repository.

With no arguments, every tracked repo that has staged changes is committed,
followed by the umbrella repository if it also has staged changes.

With one or more path arguments only the selected repositories are committed.
Arguments may be tracked repo names, repo paths, or paths inside a tracked
repo. Workspace-root paths commit the umbrella repository.`,
	Example: `  # Commit all staged repos, then the umbrella
  gyat commit -m "feat: add login endpoint"

  # Commit only specific tracked repos
  gyat commit -m "fix: typo" services/auth services/billing

  # Commit only the umbrella repository
  gyat commit -m "chore: update workspace docs" .gitignore

  # Skip git hooks
  gyat commit -m "wip" --no-verify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if commitMessage == "" {
			return fmt.Errorf("required flag \"message\" not set")
		}
		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runCommitFrom(startDir, dir, commitMessage, commitNoVerify, cmd, args)
	},
}

// hasStagedChanges reports whether git status --porcelain output contains any
// line representing a staged (index) change. In the two-column porcelain
// format "XY PATH", position 0 (X) is the index status. If X is not ' ' and
// not '?', there is a staged change.
func hasStagedChanges(statusOut string) bool {
	for _, line := range strings.Split(statusOut, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 2 {
			continue
		}
		x := line[0]
		if x != ' ' && x != '?' {
			return true
		}
	}
	return false
}

// buildCommitArgs constructs the argument slice for a git commit invocation.
func buildCommitArgs(message string, noVerify bool) []string {
	args := []string{"commit", "-m", message}
	if noVerify {
		args = append(args, "--no-verify")
	}
	return args
}

// runCommit commits staged changes in targeted repositories and/or the
// umbrella repository. The function signature mirrors runTrack so that tests
// can invoke it directly without touching global flag state.
func runCommit(dir, message string, noVerify bool, cmd *cobra.Command, args []string) error {
	return runCommitFrom(dir, dir, message, noVerify, cmd, args)
}

func runCommitFrom(startDir, dir, message string, noVerify bool, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}
	return runCommitWorkspace(ws, startDir, message, noVerify, cmd, args)
}

func runCommitWorkspace(ws workspace.Workspace, startDir, message string, noVerify bool, cmd *cobra.Command, args []string) error {
	includeRoot, repoPaths, err := resolveCommitWorkspaceTargets(ws, startDir, args)
	if err != nil {
		return err
	}

	selected := make(map[string]struct{}, len(repoPaths))
	for _, repoPath := range repoPaths {
		selected[repoPath] = struct{}{}
	}

	committed := 0
	commitArgs := buildCommitArgs(message, noVerify)

	for _, repo := range ws.Manifest.Repos {
		if _, ok := selected[repo.Path]; !ok {
			continue
		}

		repoDir := filepath.Join(ws.RootDir, filepath.FromSlash(repo.Path))
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", repo.Path)
			continue
		} else if err != nil {
			return err
		}

		statusOut, err := git.Run(repoDir, "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("checking status of '%s': %w", repo.Path, err)
		}
		if !hasStagedChanges(statusOut) {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "committing '%s'...\n", repo.Path)
		if _, err := git.Run(repoDir, commitArgs...); err != nil {
			return fmt.Errorf("committing '%s': %w", repo.Path, err)
		}
		committed++
	}

	if includeRoot {
		statusOut, err := git.Run(ws.RootDir, "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("checking umbrella status: %w", err)
		}
		if hasStagedChanges(statusOut) {
			fmt.Fprintln(cmd.ErrOrStderr(), "committing umbrella repository...")
			if _, err := git.Run(ws.RootDir, commitArgs...); err != nil {
				return fmt.Errorf("committing umbrella repository: %w", err)
			}
			committed++
		}
	}

	if committed == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to commit")
	}

	return nil
}

func resolveCommitWorkspaceTargets(ws workspace.Workspace, startDir string, args []string) (bool, []string, error) {
	if len(args) == 0 {
		repoPaths := make([]string, 0, len(ws.Manifest.Repos))
		for _, repo := range ws.Manifest.Repos {
			repoPaths = append(repoPaths, repo.Path)
		}
		return true, repoPaths, nil
	}

	includeRoot := false
	seen := make(map[string]struct{}, len(args))
	repoPaths := make([]string, 0, len(args))

	for _, arg := range args {
		selectRoot, repoPath, err := resolveCommitWorkspaceArg(ws, startDir, arg)
		if err != nil {
			return false, nil, err
		}
		if selectRoot {
			includeRoot = true
		}
		if repoPath == "" {
			continue
		}
		if _, exists := seen[repoPath]; exists {
			continue
		}
		seen[repoPath] = struct{}{}
		repoPaths = append(repoPaths, repoPath)
	}

	if !includeRoot && len(repoPaths) == 0 {
		return false, nil, fmt.Errorf("no targets selected")
	}

	return includeRoot, repoPaths, nil
}

func resolveCommitWorkspaceArg(ws workspace.Workspace, startDir, arg string) (bool, string, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return false, "", fmt.Errorf("path is required")
	}

	for _, repo := range ws.Manifest.Repos {
		if repo.Name == trimmed {
			return false, repo.Path, nil
		}
	}

	rel, err := normalizeWorkspaceArg(ws.RootDir, startDir, trimmed)
	if err != nil {
		return false, "", err
	}
	if rel == "." {
		return true, "", nil
	}

	repoPath, _, _, matched := matchTrackedRepo(ws.Manifest.Repos, rel)
	if matched {
		return false, repoPath, nil
	}

	if commitArgSelectsRoot(ws.RootDir, rel, trimmed) {
		return true, "", nil
	}

	return false, "", fmt.Errorf("'%s' is not a tracked repository or workspace path", arg)
}

func commitArgSelectsRoot(root, rel, arg string) bool {
	if filepath.IsAbs(arg) || strings.ContainsAny(arg, `/\`) || strings.HasPrefix(arg, ".") {
		return true
	}
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message (required)")
	commitCmd.Flags().BoolVar(&commitNoVerify, "no-verify", false, "Bypass git hooks")
}
