package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/git"
	"github.com/spf13/cobra"
)

var trackBranch string

var trackCmd = &cobra.Command{
	Use:   "track <repo> [path]",
	Short: "Add a repository as a submodule",
	Long: `Add a repository as a git submodule under this repository.

The <repo> argument accepts either a remote URL or a local path:

  Remote URL  - Any URL that git understands (HTTPS, SSH, git://).
  Local path  - An absolute or relative path to a repository on disk.

When using a local path, prefer a relative path (e.g. ../service-auth)
over an absolute one (e.g. /home/user/service-auth). Absolute paths are
machine-specific and will break for anyone else who clones the umbrella
repository.

Optionally specify a destination [path] inside this repository where the
submodule should live. If omitted, git will derive one from <repo>.

Use --branch to track a specific branch of the submodule.`,
	Example: `  # Remote URLs
  gyat track https://github.com/org/service-auth
  gyat track https://github.com/org/service-auth services/auth
  gyat track --branch main https://github.com/org/service-auth services/auth

  # Local paths (relative — portable)
  gyat track ../service-auth
  gyat track ../service-auth services/auth

  # Local paths (absolute — machine-specific, use with care)
  gyat track /home/user/projects/service-auth services/auth`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := execDir()
		if err != nil {
			return err
		}
		return runTrack(dir, trackBranch, cmd, args)
	},
}

// isLocalPath reports whether s looks like a local filesystem path rather than
// a remote URL. Remote URLs contain "://" (https, git, ssh, file) or use the
// SCP-style git@ syntax.
func isLocalPath(s string) bool {
	return !strings.Contains(s, "://") && !strings.HasPrefix(s, "git@")
}

// runTrack adds a repository as a submodule. branch may be empty to track the
// default branch. Accepting branch explicitly (rather than reading the
// package-level trackBranch flag) allows tests to run in parallel without
// racing on shared flag state.
func runTrack(dir, branch string, cmd *cobra.Command, args []string) error {
	source := args[0]

	// Warn when an absolute local path is used — it will not work on other machines.
	if filepath.IsAbs(source) {
		if _, err := os.Stat(source); err == nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: '%s' is an absolute path and will only work on this machine\n", source)
			fmt.Fprintf(cmd.ErrOrStderr(), "hint: use a relative path (e.g. ../%s) for portability\n", filepath.Base(source))
		}
	}

	// Git 2.38.1+ blocks local file transport by default (CVE-2022-39253).
	// Prepend -c protocol.file.allow=always so local paths work transparently.
	var gitArgs []string
	if isLocalPath(source) {
		gitArgs = []string{"-c", "protocol.file.allow=always", "submodule", "add"}
	} else {
		gitArgs = []string{"submodule", "add"}
	}

	if branch != "" {
		gitArgs = append(gitArgs, "--branch", branch)
	}

	gitArgs = append(gitArgs, source)

	if len(args) == 2 {
		gitArgs = append(gitArgs, args[1])
	}

	return git.RunInteractive(dir, gitArgs...)
}

func init() {
	trackCmd.Flags().StringVarP(&trackBranch, "branch", "b", "", "Branch of the submodule repository to track")
}
