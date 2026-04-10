package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

type repoInfo struct {
	path   string
	name   string
	url    string
	branch string
	sha    string
	status string
}

func init() {
	bindWorkspaceParallelFlag(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repositories tracked in the current gyat workspace",
	Long:  `List all repositories tracked in the current gyat workspace, along with their path, tracked branch, current commit, status, and source URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runListWithFlags(dir, sharedTargetFlags, cmd, args)
	},
}

func runList(dir string, cmd *cobra.Command, args []string) error {
	return runListWithFlags(dir, workspaceTargetFlags{}, cmd, args)
}

func runListWithFlags(dir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runListWorkspace(ws, flags, cmd)
}

func runListWorkspace(ws workspace.Workspace, flags workspaceTargetFlags, cmd *cobra.Command) error {
	stdout := cmd.OutOrStdout()
	errout := cmd.ErrOrStderr()

	if len(ws.Manifest.Repos) == 0 && !flags.rootOnly {
		fmt.Fprintln(stdout, "no managed repos found")
		fmt.Fprintln(errout, "hint: use 'gyat track <repo>' to add a repository")
		return nil
	}

	targets, err := ws.ResolveTargets(flags.targetOptions(true, nil))
	if err != nil {
		return err
	}

	results, err := workspace.RunTargets(targets, flags.runOptions(), func(target workspace.Target) (repoInfo, error) {
		if target.IsRoot {
			return collectRootInfo(ws.RootDir), nil
		}

		repo, ok := workspaceRepoByPath(ws, target.Path)
		if !ok {
			return repoInfo{}, fmt.Errorf("tracked repository '%s' not found in manifest", target.Path)
		}
		return collectRepoInfo(ws.RootDir, repo), nil
	})
	if err != nil {
		return err
	}

	repos := make([]repoInfo, 0, len(results))
	for _, result := range results {
		if !result.Ran {
			continue
		}
		if result.Err != nil {
			return result.Err
		}
		repos = append(repos, result.Value)
	}

	printRepoTable(stdout, repos)
	return nil
}

func collectRootInfo(root string) repoInfo {
	info := repoInfo{
		name: ".",
		path: ".",
		url:  "-",
	}

	rs, err := collectRepoStatus(root)
	if err != nil {
		info.branch = "(invalid)"
		info.sha = "?"
		info.status = "invalid"
		return info
	}
	info.branch = rs.branch

	sha, err := git.Run(root, "rev-parse", "--short", "HEAD")
	if err != nil || strings.TrimSpace(sha) == "" {
		info.sha = "?"
	} else {
		info.sha = sha
	}

	if len(rs.staged) == 0 && len(rs.unstaged) == 0 && len(rs.untracked) == 0 {
		info.status = "clean"
	} else {
		info.status = "modified"
	}

	return info
}

func collectRepoInfo(root string, repo manifest.Repo) repoInfo {
	info := repoInfo{
		name: repo.Name,
		path: repo.Path,
		url:  repo.URL,
	}
	if repo.Branch == "" {
		info.branch = "(default)"
	} else {
		info.branch = repo.Branch
	}

	repoDir := filepath.Join(root, filepath.FromSlash(repo.Path))
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		info.sha = "?"
		info.status = "not cloned"
		return info
	}

	sha, err := git.Run(repoDir, "rev-parse", "--short", "HEAD")
	if err != nil || strings.TrimSpace(sha) == "" {
		info.sha = "?"
	} else {
		info.sha = sha
	}

	statusOut, err := git.Run(repoDir, "status", "--porcelain")
	if err != nil {
		info.status = "invalid"
		return info
	}
	if strings.TrimSpace(statusOut) == "" {
		info.status = "clean"
		return info
	}
	info.status = "modified"
	return info
}

func printRepoTable(out io.Writer, repos []repoInfo) {
	const colPad = 3

	headers := [5]string{"PATH", "BRANCH", "COMMIT", "STATUS", "URL"}
	widths := [5]int{}
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, repo := range repos {
		vals := [5]string{repo.path, repo.branch, repo.sha, repo.status, repo.url}
		for i, v := range vals {
			if len(v) > widths[i] {
				widths[i] = len(v)
			}
		}
	}

	// formatRow formats all columns with consistent widths.
	// The last column (URL) is not right-padded — it is always the final field.
	formatRow := func(cols [5]string) string {
		return fmt.Sprintf(
			"%-*s%-*s%-*s%-*s%s",
			widths[0]+colPad, cols[0],
			widths[1]+colPad, cols[1],
			widths[2]+colPad, cols[2],
			widths[3]+colPad, cols[3],
			cols[4],
		)
	}

	separatorWidth := widths[0] + widths[1] + widths[2] + widths[3] + widths[4] + colPad*4

	fmt.Fprintln(out, formatRow(headers))
	fmt.Fprintln(out, strings.Repeat("-", separatorWidth))
	for _, repo := range repos {
		fmt.Fprintln(out, formatRow([5]string{repo.path, repo.branch, repo.sha, repo.status, repo.url}))
	}
}
