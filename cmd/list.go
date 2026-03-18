package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

type submoduleInfo struct {
	path   string
	url    string
	branch string
	sha    string
	status string
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed submodules",
	Long:  `List all submodules tracked in this repository, along with their URL, tracked branch, current commit, and status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		return runList(dir, cmd, args)
	},
}

func runList(dir string, cmd *cobra.Command, args []string) error {
	stdout := cmd.OutOrStdout()
	errout := cmd.ErrOrStderr()

	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); os.IsNotExist(err) {
		fmt.Fprintln(stdout, "no submodules found")
		fmt.Fprintln(errout, "hint: use 'gyat add <repo>' to add a repository")
		return nil
	}

	// Gather all submodule paths registered in .gitmodules.
	pathsOut, err := git.Run(dir, "config", "-f", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil || strings.TrimSpace(pathsOut) == "" {
		fmt.Fprintln(stdout, "no submodules found")
		fmt.Fprintln(errout, "hint: use 'gyat add <repo>' to add a repository")
		return nil
	}

	// Parse "submodule.<name>.path <path>" lines into a name->info map.
	submodules := []submoduleInfo{}
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
		submodules = append(submodules, submoduleInfo{path: path})
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

	// Print the table.
	// Two-pass layout: measure actual column widths from the data first, then
	// format every row with consistent %-*s padding.
	const colPad = 3

	headers := [5]string{"PATH", "BRANCH", "COMMIT", "STATUS", "URL"}
	widths := [5]int{}
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, sub := range submodules {
		vals := [5]string{sub.path, sub.branch, sub.sha, sub.status, sub.url}
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

	fmt.Fprintln(stdout, formatRow(headers))
	fmt.Fprintln(stdout, strings.Repeat("-", separatorWidth))
	for _, sub := range submodules {
		fmt.Fprintln(stdout, formatRow([5]string{sub.path, sub.branch, sub.sha, sub.status, sub.url}))
	}

	return nil
}
