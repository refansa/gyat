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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repositories tracked in the current gyat workspace",
	Long:  `List all repositories tracked in the current gyat workspace, along with their path, tracked branch, current commit, status, and source URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runList(dir, cmd, args)
	},
}

func runList(dir string, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runListWorkspace(ws, cmd)
}

func runListWorkspace(ws workspace.Workspace, cmd *cobra.Command) error {
	stdout := cmd.OutOrStdout()
	errout := cmd.ErrOrStderr()

	if len(ws.Manifest.Repos) == 0 {
		fmt.Fprintln(stdout, "no managed repos found")
		fmt.Fprintln(errout, "hint: use 'gyat track <repo>' to add a repository")
		return nil
	}

	repos := make([]repoInfo, 0, len(ws.Manifest.Repos))
	for _, repo := range ws.Manifest.Repos {
		repos = append(repos, collectRepoInfo(ws.RootDir, repo))
	}

	printRepoTable(stdout, repos)
	return nil
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
