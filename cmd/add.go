package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path...]",
	Short: "Stage changes across the umbrella repository and submodules",
	Long: `Stage changes in the umbrella repository and any registered submodules.

Behaves like 'git add' but is submodule-aware: paths that live inside a
submodule are routed to that submodule's index, while all other paths are
staged in the umbrella repository itself.

With no arguments, everything is staged:
  - All working-tree changes in the umbrella root (equivalent to git add -A)
  - All working-tree changes inside every checked-out submodule

With one or more path arguments each path is routed to the repository it
belongs to:
  - A submodule path (e.g. services/auth)          → git add -A inside that submodule
  - A path inside a submodule (e.g. services/auth/handler.go) → git add <file> inside the submodule
  - Any other path (e.g. .gitignore, README.md)    → staged in the umbrella root`,
	Example: `  # Stage everything: umbrella root + all submodules
  gyat add

  # Stage a specific file in the umbrella root
  gyat add .gitignore

  # Stage all changes inside a submodule
  gyat add services/auth

  # Stage a specific file inside a submodule
  gyat add services/auth/handler.go

  # Mix of root files and submodule paths
  gyat add README.md services/auth services/billing/main.go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		return runAdd(dir, cmd, args)
	},
}

// subStage describes what to stage inside a specific submodule.
// When stageAll is true, git add -A is used; otherwise only the listed files
// are staged.
type subStage struct {
	stageAll bool
	files    []string
}

// runAdd stages changes in the umbrella repository and/or its submodules.
// With no args it stages everything (root + all submodules). With args it
// routes each path to the repository it belongs to.
func runAdd(dir string, cmd *cobra.Command, args []string) error {
	submodulePaths, err := allSubmodulePaths(dir)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return stageAll(dir, submodulePaths, cmd)
	}
	return stageTargeted(dir, submodulePaths, args, cmd)
}

// stageAll runs git add -A in the umbrella root and then inside every
// checked-out submodule that has working-tree changes.
func stageAll(dir string, submodulePaths []string, cmd *cobra.Command) error {
	staged := 0

	// Umbrella root.
	statusOut, err := git.Run(dir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("checking umbrella status: %w", err)
	}
	if hasWorkingTreeChanges(statusOut) {
		fmt.Fprintln(cmd.ErrOrStderr(), "staging umbrella repository...")
		if _, err := git.Run(dir, "add", "-A"); err != nil {
			return fmt.Errorf("staging umbrella root: %w", err)
		}
		staged++
	}

	// Submodules.
	for _, path := range submodulePaths {
		subDir := filepath.Join(dir, path)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", path)
			continue
		}

		statusOut, err := git.Run(subDir, "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("checking status of '%s': %w", path, err)
		}
		if !hasWorkingTreeChanges(statusOut) {
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "staging '%s'...\n", path)
		if _, err := git.Run(subDir, "add", "-A"); err != nil {
			return fmt.Errorf("staging '%s': %w", path, err)
		}
		staged++
	}

	if staged == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing to stage")
	}

	return nil
}

// stageTargeted classifies each argument by its owning repository and stages
// only the requested paths.
func stageTargeted(dir string, submodulePaths []string, args []string, cmd *cobra.Command) error {
	rootArgs, subTargets := classifyArgs(submodulePaths, args)

	if len(rootArgs) > 0 {
		gitArgs := append([]string{"add", "--"}, rootArgs...)
		if _, err := git.Run(dir, gitArgs...); err != nil {
			return fmt.Errorf("staging root paths: %w", err)
		}
	}

	for sub, stage := range subTargets {
		subDir := filepath.Join(dir, sub)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: submodule '%s' is not checked out, skipping\n", sub)
			continue
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "staging '%s'...\n", sub)

		var gitArgs []string
		if stage.stageAll {
			gitArgs = []string{"add", "-A"}
		} else {
			gitArgs = append([]string{"add", "--"}, stage.files...)
		}

		if _, err := git.Run(subDir, gitArgs...); err != nil {
			return fmt.Errorf("staging '%s': %w", sub, err)
		}
	}

	return nil
}

// classifyArgs routes each argument to the repository it belongs to.
// rootArgs receives paths that do not fall inside any registered submodule.
// subTargets maps each matched submodule path to the files to stage inside it;
// stageAll true means "git add -A" (triggered when the arg is the submodule
// root itself), otherwise only the specific relative paths are staged.
func classifyArgs(submodulePaths []string, args []string) (rootArgs []string, subTargets map[string]*subStage) {
	subTargets = make(map[string]*subStage)

	for _, arg := range args {
		norm := filepath.ToSlash(filepath.Clean(arg))
		matched := false

		for _, sub := range submodulePaths {
			subNorm := filepath.ToSlash(filepath.Clean(sub))

			if norm == subNorm {
				// Exact submodule match — stage everything inside it.
				subTargets[sub] = &subStage{stageAll: true}
				matched = true
				break
			}

			if strings.HasPrefix(norm, subNorm+"/") {
				rel := strings.TrimPrefix(norm, subNorm+"/")
				if t, exists := subTargets[sub]; exists {
					// If we're already staging all, this file is already covered.
					if !t.stageAll {
						t.files = append(t.files, rel)
					}
				} else {
					subTargets[sub] = &subStage{files: []string{rel}}
				}
				matched = true
				break
			}
		}

		if !matched {
			rootArgs = append(rootArgs, arg)
		}
	}

	return
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

// allSubmodulePaths reads every submodule path from .gitmodules.
// Returns nil (not an error) when .gitmodules is absent or empty.
func allSubmodulePaths(dir string) ([]string, error) {
	pathsOut, err := git.Run(dir, "config", "-f", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil || strings.TrimSpace(pathsOut) == "" {
		return nil, nil
	}

	var paths []string
	for _, line := range strings.Split(pathsOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		paths = append(paths, parts[1])
	}
	return paths, nil
}
