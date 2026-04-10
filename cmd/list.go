package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/refansa/gyat/internal/manifest"
	"github.com/refansa/gyat/internal/workspace"
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
	if err == nil {
		return runListWorkspace(ws, cmd)
	}
	if !errors.Is(err, workspace.ErrNotFound) {
		return err
	}

	return runListLegacy(dir, cmd, args)
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

func runListLegacy(dir string, cmd *cobra.Command, args []string) error {
	stdout := cmd.OutOrStdout()
	errout := cmd.ErrOrStderr()

	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); os.IsNotExist(err) {
		fmt.Fprintln(stdout, "no submodules found")
		fmt.Fprintln(errout, "hint: use 'gyat track <repo>' to add a repository")
		return nil
	}

	// Gather all submodule paths registered in .gitmodules.
	pathsOut, err := git.Run(dir, "config", "-f", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil || strings.TrimSpace(pathsOut) == "" {
		fmt.Fprintln(stdout, "no submodules found")
		fmt.Fprintln(errout, "hint: use 'gyat track <repo>' to add a repository")
		return nil
	}

	// Parse "submodule.<name>.path <path>" lines into a name->info map.
	submodules := []repoInfo{}
	nameByPath := map[string]string{}

	for _, line := range strings.Split(pathsOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		// key is like "submodule.services/auth.path"
		key := parts[0]
		path := parts[1]

		// Extract the submodule name from between the first and last dots.
		withoutPrefix := strings.TrimPrefix(key, "submodule.")
		name := strings.TrimSuffix(withoutPrefix, ".path")

		nameByPath[path] = name
		submodules = append(submodules, repoInfo{path: path})
	}

	// Enrich each submodule with its URL and branch from .gitmodules.
	for i, sub := range submodules {
		name := nameByPath[sub.path]

		url, _ := git.Run(dir, "config", "-f", ".gitmodules", fmt.Sprintf("submodule.%s.url", name))
		branch, _ := git.Run(dir, "config", "-f", ".gitmodules", fmt.Sprintf("submodule.%s.branch", name))

		submodules[i].url = url
		if branch == "" {
			submodules[i].branch = "(default)"
		} else {
			submodules[i].branch = branch
		}
	}

	// Parse "git submodule status" for SHA and status.
	// Output format: [prefix]<sha1> <path> [(<describe>)]
	// Prefix: ' ' = ok, '-' = not initialized, '+' = modified, 'U' = conflict
	statusOut, _ := git.Run(dir, "submodule", "status")
	statusByPath := map[string]struct {
		sha    string
		status string
	}{}

	for _, line := range strings.Split(statusOut, "\n") {
		if len(line) < 2 {
			continue
		}

		prefix := line[0]
		rest := strings.Fields(line[1:])
		if len(rest) < 2 {
			continue
		}

		sha := rest[0]
		if len(sha) > 8 {
			sha = sha[:8]
		}
		path := rest[1]

		var label string
		switch prefix {
		case '-':
			label = "not initialized"
		case '+':
			label = "modified"
		case 'U':
			label = "merge conflict"
		default:
			label = "up to date"
		}

		statusByPath[path] = struct {
			sha    string
			status string
		}{sha: sha, status: label}
	}

	for i, sub := range submodules {
		if info, ok := statusByPath[sub.path]; ok {
			submodules[i].sha = info.sha
			submodules[i].status = info.status
		} else {
			submodules[i].sha = "?"
			submodules[i].status = "unknown"
		}
	}

	printRepoTable(stdout, submodules)

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
