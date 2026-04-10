package cmd

import (
	"bytes"
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

In interactive terminals, the report is paged automatically. Use '--no-pager'
to print directly to stdout instead.

Without selector flags, status shows the umbrella repository followed by every
tracked repo. Use positional repo names or paths, '--repo', and '--group' to
narrow the repo set, '--no-root' to exclude the umbrella repository, or
'--root-only' to inspect only the umbrella root.`,
	Example: `  # Show status for all repositories
  gyat status

	# Show status for specific repos while keeping the umbrella
	gyat status services/auth services/billing

	# Show status for one repo without the umbrella
	gyat status --repo auth --no-root

	# Show only the umbrella repository
	gyat status --root-only

	# Print directly without a pager
	gyat status --no-pager

  # Show status for a repo selected by name
  gyat status auth`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runStatusWithFlags(dir, sharedTargetFlags, cmd, args)
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

type statusTargetResult struct {
	label       string
	unavailable string
	status      repoStatus
}

func init() {
	bindWorkspaceParallelFlag(statusCmd)
	bindNoPagerFlag(statusCmd)
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
	return runStatusWithFlags(dir, workspaceTargetFlags{}, cmd, args)
}

func runStatusWithFlags(dir string, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runStatusWorkspace(ws, flags, cmd, args)
}

// runStatusWorkspace prints a status report for the selected workspace targets.
// Positional args are treated as repo selectors (name or path) and are
// combined with the shared target flags.
func runStatusWorkspace(ws workspace.Workspace, flags workspaceTargetFlags, cmd *cobra.Command, args []string) error {
	stdout := cmd.OutOrStdout()
	errout := cmd.ErrOrStderr()
	var report bytes.Buffer
	var failures commandFailures

	targets, err := ws.ResolveTargets(flags.targetOptions(true, args))
	if err != nil {
		return err
	}

	results, err := workspace.RunTargets(targets, flags.runOptions(), func(target workspace.Target) (statusTargetResult, error) {
		label := target.Path
		if target.IsRoot {
			label = "umbrella repository"
		} else if _, statErr := os.Stat(target.Dir); os.IsNotExist(statErr) {
			return statusTargetResult{label: target.Path, unavailable: "not cloned"}, nil
		} else if statErr != nil {
			return statusTargetResult{label: label}, fmt.Errorf("stat '%s': %w", target.Label, statErr)
		}

		rs, err := collectRepoStatus(target.Dir)
		if err != nil {
			return statusTargetResult{label: label}, fmt.Errorf("could not get status of '%s': %w", target.Label, err)
		}

		return statusTargetResult{label: label, status: rs}, nil
	})
	if err != nil {
		return err
	}

	for _, result := range results {
		if !result.Ran {
			continue
		}
		if result.Err != nil {
			if handledErr := failures.handleErr(flags.continueOnError, result.Err); handledErr != nil {
				return handledErr
			}
			fmt.Fprintf(errout, "warning: %v\n", result.Err)
			continue
		}

		if result.Value.unavailable != "" {
			printUnavailableRepo(&report, result.Value.label, result.Value.unavailable)
			continue
		}

		printRepoSection(&report, result.Value.label, result.Value.status)
	}

	if err := writeMaybePagedOutput(stdout, errout, report.String(), noPagerEnabled(cmd)); err != nil {
		return err
	}

	if len(ws.Manifest.Repos) == 0 && !flags.rootOnly && !flags.noRoot {
		fmt.Fprintln(errout, "hint: use 'gyat track <repo>' to add a repository")
	}

	return failures.err("status failed")
}
