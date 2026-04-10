package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path...]",
	Short: "Stage changes across the umbrella repository and tracked repos",
	Long: `Stage changes in the umbrella repository and tracked repos.

Within a .gyat workspace, paths are resolved relative to the current working
directory and routed to the repository that owns them. Paths that live inside a
tracked repo are staged inside that repo, while all other paths are staged in
the umbrella repository itself.

With no arguments, everything is staged:
  - All working-tree changes in the umbrella root (equivalent to git add -A)
	- All working-tree changes inside every cloned tracked repo

With one or more path arguments each path is routed to the repository it
belongs to:
	- A tracked repo path (e.g. services/auth)       → git add -A inside that repo
	- A path inside a tracked repo (e.g. services/auth/handler.go) → git add <file> inside that repo
  - Any other path (e.g. .gitignore, README.md)    → staged in the umbrella root`,
	Example: `  # Stage everything: umbrella root + all tracked repos
  gyat add

  # Stage a specific file in the umbrella root
  gyat add .gitignore

	# Stage all changes inside a tracked repo
  gyat add services/auth

	# Stage a specific file inside a tracked repo
  gyat add services/auth/handler.go

	# Mix of root files and repo paths
  gyat add README.md services/auth services/billing/main.go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}

		dir, err := execDir()
		if err != nil {
			return err
		}
		return runAddWithFlagsFrom(startDir, dir, sharedTargetFlags, cmd, args)
	},
}

// repoStage describes what to stage inside a specific tracked repository.
// When stageAll is true, git add -A is used; otherwise only the listed files
// are staged.
type repoStage struct {
	stageAll bool
	files    []string
}

// runAdd stages changes in the umbrella repository and/or tracked repos.
// With no args it stages everything in the workspace. With args it routes each
// path to the repository it belongs to.
func runAdd(dir string, cmd *cobra.Command, args []string) error {
	return runAddWithFlagsFrom(dir, dir, workspaceTargetFlags{}, cmd, args)
}

func runAddFrom(startDir, dir string, cmd *cobra.Command, args []string) error {
	return runAddWithFlagsFrom(startDir, dir, workspaceTargetFlags{}, cmd, args)
}

func runAddWithFlagsFrom(startDir, dir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	_ = dir
	ws, err := workspace.Load(startDir)
	if err != nil {
		return err
	}
	return runAddWorkspace(ws, startDir, flags, cmd, args)
}

func runAddWorkspace(ws workspace.Workspace, startDir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	if flags.hasSelection() {
		return stageSelectedWorkspace(ws, flags, cmd, args)
	}
	if len(args) == 0 {
		return stageAllWorkspace(ws, flags.continueOnError, cmd)
	}
	return stageTargetedWorkspace(ws, startDir, args, flags.continueOnError, cmd)
}

func stageAllWorkspace(ws workspace.Workspace, continueOnError bool, cmd *cobra.Command) error {
	staged := 0
	var failures commandFailures

	statusOut, err := git.Run(ws.RootDir, "status", "--porcelain")
	if err != nil {
		if handledErr := failures.handle(continueOnError, "checking umbrella status: %w", err); handledErr != nil {
			return handledErr
		}
	} else if hasWorkingTreeChanges(statusOut) {
		fmt.Fprintln(cmd.ErrOrStderr(), "staging umbrella repository...")
		if _, err := git.Run(ws.RootDir, "add", "-A"); err != nil {
			if handledErr := failures.handle(continueOnError, "staging umbrella root: %w", err); handledErr != nil {
				return handledErr
			}
		} else {
			staged++
		}
	}

	for _, repo := range ws.Manifest.Repos {
		repoDir := filepath.Join(ws.RootDir, filepath.FromSlash(repo.Path))
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", repo.Path)
			continue
		} else if err != nil {
			return err
		}

		statusOut, err := git.Run(repoDir, "status", "--porcelain")
		if err != nil {
			if handledErr := failures.handle(continueOnError, "checking status of '%s': %w", repo.Path, err); handledErr != nil {
				return handledErr
			}
			continue
		}
		if !hasWorkingTreeChanges(statusOut) {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "staging '%s'...\n", repo.Path)
		if _, err := git.Run(repoDir, "add", "-A"); err != nil {
			if handledErr := failures.handle(continueOnError, "staging '%s': %w", repo.Path, err); handledErr != nil {
				return handledErr
			}
			continue
		}
		staged++
	}

	if staged == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to stage")
	}

	return failures.err("staging failed")
}

func stageSelectedWorkspace(ws workspace.Workspace, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	targets, err := ws.ResolveTargets(flags.targetOptions(true, nil))
	if err != nil {
		return err
	}

	staged := 0
	var failures commandFailures
	for _, target := range targets {
		if !target.IsRoot {
			if _, err := os.Stat(target.Dir); os.IsNotExist(err) {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", target.Path)
				continue
			} else if err != nil {
				if handledErr := failures.handle(flags.continueOnError, "checking target '%s': %w", target.Label, err); handledErr != nil {
					return handledErr
				}
				continue
			}
		}

		label := target.Label
		if target.IsRoot {
			label = "umbrella repository"
		}

		gitArgs := []string{"add", "-A"}
		if len(args) > 0 {
			gitArgs = append([]string{"add", "--"}, args...)
		}
		if len(args) == 0 {
			statusOut, err := git.Run(target.Dir, "status", "--porcelain")
			if err != nil {
				if handledErr := failures.handle(flags.continueOnError, "checking status of '%s': %w", label, err); handledErr != nil {
					return handledErr
				}
				continue
			}
			if !hasWorkingTreeChanges(statusOut) {
				continue
			}
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "staging '%s'...\n", label)
		if _, err := git.Run(target.Dir, gitArgs...); err != nil {
			if handledErr := failures.handle(flags.continueOnError, "staging '%s': %w", label, err); handledErr != nil {
				return handledErr
			}
			continue
		}
		staged++
	}

	if staged == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to stage")
	}

	return failures.err("staging failed")
}

func stageTargetedWorkspace(ws workspace.Workspace, startDir string, args []string, continueOnError bool, cmd *cobra.Command) error {
	rootArgs, repoTargets, err := classifyWorkspaceArgs(ws.RootDir, ws.Manifest.Repos, startDir, args)
	if err != nil {
		return err
	}
	var failures commandFailures

	if len(rootArgs) > 0 {
		gitArgs := append([]string{"add", "--"}, rootArgs...)
		if _, err := git.Run(ws.RootDir, gitArgs...); err != nil {
			if handledErr := failures.handle(continueOnError, "staging root paths: %w", err); handledErr != nil {
				return handledErr
			}
		}
	}

	for _, repo := range ws.Manifest.Repos {
		stage, ok := repoTargets[repo.Path]
		if !ok {
			continue
		}

		repoDir := filepath.Join(ws.RootDir, filepath.FromSlash(repo.Path))
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: tracked repository '%s' is not cloned, skipping\n", repo.Path)
			continue
		} else if err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "staging '%s'...\n", repo.Path)

		var gitArgs []string
		if stage.stageAll {
			gitArgs = []string{"add", "-A"}
		} else {
			gitArgs = append([]string{"add", "--"}, stage.files...)
		}

		if _, err := git.Run(repoDir, gitArgs...); err != nil {
			if handledErr := failures.handle(continueOnError, "staging '%s': %w", repo.Path, err); handledErr != nil {
				return handledErr
			}
		}
	}

	return failures.err("staging failed")
}

func classifyWorkspaceArgs(root string, repos []manifest.Repo, startDir string, args []string) (rootArgs []string, repoTargets map[string]*repoStage, err error) {
	repoTargets = make(map[string]*repoStage)

	for _, arg := range args {
		rel, err := normalizeWorkspaceArg(root, startDir, arg)
		if err != nil {
			return nil, nil, err
		}

		repoPath, repoArg, stageAll, matched := matchTrackedRepo(repos, rel)
		if !matched {
			rootArgs = append(rootArgs, rel)
			continue
		}

		if stage, exists := repoTargets[repoPath]; exists {
			if stage.stageAll {
				continue
			}
			if stageAll {
				repoTargets[repoPath] = &repoStage{stageAll: true}
				continue
			}
			stage.files = appendUnique(stage.files, repoArg)
			continue
		}

		stage := &repoStage{stageAll: stageAll}
		if !stageAll {
			stage.files = []string{repoArg}
		}
		repoTargets[repoPath] = stage
	}

	return rootArgs, repoTargets, nil
}

func normalizeWorkspaceArg(root, startDir, arg string) (string, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", fmt.Errorf("path is required")
	}

	resolved := arg
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(startDir, filepath.FromSlash(resolved))
	}

	absPath, err := filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve path '%s': %w", arg, err)
	}

	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return "", fmt.Errorf("resolve path '%s': %w", arg, err)
	}

	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("path '%s' must stay within the workspace", arg)
	}

	return rel, nil
}

func matchTrackedRepo(repos []manifest.Repo, arg string) (repoPath, repoArg string, stageAll bool, matched bool) {
	bestLen := -1

	for _, repo := range repos {
		switch {
		case arg == repo.Path:
			if len(repo.Path) > bestLen {
				repoPath = repo.Path
				repoArg = ""
				stageAll = true
				matched = true
				bestLen = len(repo.Path)
			}
		case strings.HasPrefix(arg, repo.Path+"/"):
			if len(repo.Path) > bestLen {
				repoPath = repo.Path
				repoArg = strings.TrimPrefix(arg, repo.Path+"/")
				stageAll = false
				matched = true
				bestLen = len(repo.Path)
			}
		}
	}

	return repoPath, repoArg, stageAll, matched
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

// hasWorkingTreeChanges reports whether git status --porcelain output contains
// any line representing an unstaged working-tree change. In the porcelain
// format each line is "XY PATH" where X is the index status and Y is the
// working-tree status. Lines where Y is ' ' represent index-only changes that
// are already staged — those do not need another git add.
func hasWorkingTreeChanges(statusOut string) bool {
	for _, line := range strings.Split(statusOut, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 2 {
			continue
		}
		// Y column (position 1): ' ' means working tree is clean for this entry.
		// '?' means untracked, 'M'/'D'/etc mean modified/deleted in working tree.
		if line[1] != ' ' {
			return true
		}
	}
	return false
}
