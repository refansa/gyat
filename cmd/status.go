package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [path...]",
	Short: "Show working tree status across the umbrella repository and tracked repos",
	Long: `Show the working tree status of the umbrella repository and tracked repos
in the current gyat workspace.

For each repository a section is printed that mirrors 'git status': staged
changes, unstaged changes, and untracked files are listed under clearly
labelled headings. Repositories listed in .gyat but missing on disk are flagged
with "not cloned".

Without arguments every tracked repo is shown alongside the umbrella
repository. Pass one or more repo names or paths to limit the output to those
repos (the umbrella is always shown).`,
	Example: `  # Show status for all repositories
  gyat status

  # Show status for specific repos (plus the umbrella)
  gyat status services/auth services/billing

  # Show status for a repo selected by name
  gyat status auth`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runStatus(dir, cmd, args)
	},
}

// statusEntry holds a single file's index (X) and worktree (Y) status codes
// together with the path reported by "git status --porcelain".
type statusEntry struct {
	x    byte
	y    byte
	path string
}

// repoStatus holds the parsed status for one repository.
type repoStatus struct {
	branch    string
	staged    []statusEntry
	unstaged  []statusEntry
	untracked []statusEntry
}

// parsePorcelain parses the output of "git status --porcelain" into a slice of
// statusEntry values. Lines shorter than four bytes (minimum "XY PATH") are
// silently skipped.
func parsePorcelain(out string) []statusEntry {
	var entries []statusEntry
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}
		entries = append(entries, statusEntry{
			x:    line[0],
			y:    line[1],
			path: line[3:],
		})
	}
	return entries
}

// statusLabel returns the human-readable label for a single porcelain status
// code byte. The returned string does not include a trailing colon.
func statusLabel(code byte) string {
	switch code {
	case 'A':
		return "new file"
	case 'M':
		return "modified"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	case 'C':
		return "copied"
	case 'U':
		return "conflict"
	default:
		return "changed"
	}
}

// collectRepoStatus retrieves the branch name and parses "git status
// --porcelain" for the repository at dir, classifying each entry as staged,
// unstaged, or untracked.
//
// symbolic-ref is preferred over branch --show-current so that the branch name
// is available even in a repository that has not yet produced any commits.
// When symbolic-ref fails the repository is in detached-HEAD state.
func collectRepoStatus(dir string) (repoStatus, error) {
	var rs repoStatus

	branch, err := git.Run(dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		sha, shaErr := git.Run(dir, "rev-parse", "--short", "HEAD")
		if shaErr == nil && sha != "" {
			rs.branch = "HEAD detached at " + sha
		} else {
			rs.branch = "(no branch)"
		}
	} else {
		rs.branch = strings.TrimSpace(branch)
	}

	statusOut, err := git.Run(dir, "status", "--porcelain")
	if err != nil {
		return rs, fmt.Errorf("git status: %w", err)
	}

	for _, e := range parsePorcelain(statusOut) {
		switch {
		case e.x == '?' && e.y == '?':
			rs.untracked = append(rs.untracked, e)
		default:
			if e.x != ' ' && e.x != '?' {
				rs.staged = append(rs.staged, e)
			}
			if e.y != ' ' && e.y != '?' {
				rs.unstaged = append(rs.unstaged, e)
			}
		}
	}

	return rs, nil
}

// printRepoSection writes one repository's status block to out.
// label is the human-friendly name shown in the section header (e.g.
// "umbrella repository" or a tracked repo path like "services/auth").
func printRepoSection(out io.Writer, label string, rs repoStatus) {
	header := fmt.Sprintf("%s — %s", label, rs.branch)
	sep := strings.Repeat("─", utf8.RuneCountInString(header))
	fmt.Fprintln(out, header)
	fmt.Fprintln(out, sep)

	if len(rs.staged) == 0 && len(rs.unstaged) == 0 && len(rs.untracked) == 0 {
		fmt.Fprintln(out, "nothing to commit, working tree clean")
		fmt.Fprintln(out)
		return
	}

	if len(rs.staged) > 0 {
		fmt.Fprintln(out, "Changes to be committed:")
		for _, e := range rs.staged {
			fmt.Fprintf(out, "\t%-12s%s\n", statusLabel(e.x)+":", e.path)
		}
		fmt.Fprintln(out)
	}

	if len(rs.unstaged) > 0 {
		fmt.Fprintln(out, "Changes not staged for commit:")
		for _, e := range rs.unstaged {
			fmt.Fprintf(out, "\t%-12s%s\n", statusLabel(e.y)+":", e.path)
		}
		fmt.Fprintln(out)
	}

	if len(rs.untracked) > 0 {
		fmt.Fprintln(out, "Untracked files:")
		for _, e := range rs.untracked {
			fmt.Fprintf(out, "\t%s\n", e.path)
		}
		fmt.Fprintln(out)
	}
}

// printUnavailableRepo writes a status block to out for a tracked repository
// whose working-tree directory is not available on disk.
func printUnavailableRepo(out io.Writer, path, status string) {
	header := fmt.Sprintf("%s — (%s)", path, status)
	sep := strings.Repeat("─", utf8.RuneCountInString(header))
	fmt.Fprintln(out, header)
	fmt.Fprintln(out, sep)
	fmt.Fprintln(out, status)
	fmt.Fprintln(out)
}

func runStatus(dir string, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runStatusWorkspace(ws, cmd, args)
}

// runStatusWorkspace prints a status report for the umbrella repository and the
// tracked repos listed in .gyat. When args is non-empty they are treated as
// repo selectors (name or path), while the umbrella repository is always shown.
func runStatusWorkspace(ws workspace.Workspace, cmd *cobra.Command, args []string) error {
	stdout := cmd.OutOrStdout()
	errout := cmd.ErrOrStderr()

	// Always show the umbrella repository first.
	umbrellaStatus, err := collectRepoStatus(ws.RootDir)
	if err != nil {
		return fmt.Errorf("umbrella repository: %w", err)
	}
	printRepoSection(stdout, "umbrella repository", umbrellaStatus)

	if len(ws.Manifest.Repos) == 0 {
		fmt.Fprintln(errout, "hint: use 'gyat track <repo>' to add a repository")
		return nil
	}

	targets, err := ws.ResolveTargets(workspace.TargetOptions{
		IncludeRoot:   true,
		RepoSelectors: args,
	})
	if err != nil {
		return err
	}

	for _, target := range targets {
		if target.IsRoot {
			continue
		}

		if _, statErr := os.Stat(target.Dir); os.IsNotExist(statErr) {
			printUnavailableRepo(stdout, target.Path, "not cloned")
			continue
		}

		rs, err := collectRepoStatus(target.Dir)
		if err != nil {
			fmt.Fprintf(errout, "warning: could not get status of '%s': %v\n", target.Path, err)
			continue
		}
		printRepoSection(stdout, target.Path, rs)
	}

	return nil
}
