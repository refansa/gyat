package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var trackBranch string

var trackCmd = &cobra.Command{
	Use:   "track <repo> [path]",
	Short: "Clone and register a repository in the current gyat workspace",
	Long: `Clone a repository into the current gyat workspace and register it in
the .gyat manifest.

The <repo> argument accepts either a remote URL or a local path:

  Remote URL  - Any URL that git understands (HTTPS, SSH, git://).
  Local path  - An absolute or relative path to a repository on disk.

When using a local path, prefer a relative path (e.g. ../service-auth)
over an absolute one (e.g. /home/user/service-auth). Absolute paths are
machine-specific and will break for anyone else who clones the umbrella
repository.

Optionally specify a destination [path] inside this repository where the
managed repo should live. If omitted, gyat derives one from <repo>.

Use --branch to clone a specific branch and record it in the manifest.`,
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
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
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

// runTrack registers a repository in the current gyat workspace.
func runTrack(dir, branch string, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}
	return runTrackWorkspace(ws, dir, branch, cmd, args)
}

func runTrackWorkspace(ws workspace.Workspace, startDir, branch string, cmd *cobra.Command, args []string) error {
	source := args[0]
	destination := deriveTrackedPath(source)
	if len(args) == 2 {
		destination = normalizeTrackedPath(args[1])
	}
	if err := validateTrackedPath(destination); err != nil {
		return err
	}

	name := path.Base(destination)
	if err := validateTrackedRepo(ws.Manifest, name, destination); err != nil {
		return err
	}

	destDir := filepath.Join(ws.RootDir, filepath.FromSlash(destination))
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("destination '%s' already exists", destination)
	} else if !os.IsNotExist(err) {
		return err
	}

	cloneSource, storedURL, err := resolveTrackSource(startDir, ws.RootDir, source, cmd)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "cloning '%s' into '%s'...\n", source, destination)
	cloneArgs := []string{"clone"}
	if branch != "" {
		cloneArgs = append(cloneArgs, "--branch", branch, "--single-branch")
	}
	cloneArgs = append(cloneArgs, cloneSource, destDir)
	if _, err := git.Run(ws.RootDir, cloneArgs...); err != nil {
		return fmt.Errorf("cloning '%s': %w", source, err)
	}

	updated := ws.Manifest
	repo := manifest.Repo{
		Name: name,
		Path: destination,
		URL:  storedURL,
	}
	if branch != "" {
		repo.Branch = branch
	}
	updated.Repos = append(updated.Repos, repo)

	if err := manifest.SaveDir(ws.RootDir, updated); err != nil {
		return err
	}
	if changed, err := workspace.SyncGitIgnore(ws.RootDir, updated); err != nil {
		return err
	} else if changed {
		fmt.Fprintln(cmd.ErrOrStderr(), "updated .gitignore managed block")
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "tracked repository '%s'\n", destination)
	fmt.Fprintln(cmd.ErrOrStderr(), "hint: commit the changes to .gyat and .gitignore")
	return nil
}

func resolveTrackSource(startDir, workspaceRoot, source string, cmd *cobra.Command) (cloneSource string, storedURL string, err error) {
	if !isLocalPath(source) {
		return source, strings.TrimSpace(source), nil
	}

	if filepath.IsAbs(source) {
		if _, statErr := os.Stat(source); statErr == nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: '%s' is an absolute path and will only work on this machine\n", source)
			fmt.Fprintf(cmd.ErrOrStderr(), "hint: use a relative path (e.g. ../%s) for portability\n", filepath.Base(source))
		}
	}

	resolved, err := resolveLocalSource(startDir, source)
	if err != nil {
		return "", "", err
	}
	if _, err := git.Run(resolved, "rev-parse", "--is-inside-work-tree"); err != nil {
		return "", "", fmt.Errorf("'%s' is not a git repository", source)
	}

	storedURL = strings.TrimSpace(source)
	if remoteURL, remoteErr := git.Run(resolved, "config", "--get", "remote.origin.url"); remoteErr == nil && strings.TrimSpace(remoteURL) != "" {
		storedURL = remoteURL
	} else if relative, relErr := filepath.Rel(workspaceRoot, resolved); relErr == nil {
		storedURL = filepath.ToSlash(relative)
	} else {
		storedURL = resolved
	}

	return resolved, storedURL, nil
}

func resolveLocalSource(startDir, source string) (string, error) {
	if filepath.IsAbs(source) {
		if _, err := os.Stat(source); err != nil {
			return "", err
		}
		return filepath.Clean(source), nil
	}

	resolved := filepath.Join(startDir, filepath.FromSlash(source))
	if _, err := os.Stat(resolved); err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

func deriveTrackedPath(source string) string {
	if isLocalPath(source) {
		return normalizeTrackedPath(filepath.Base(filepath.Clean(filepath.FromSlash(source))))
	}

	trimmed := strings.TrimSpace(source)
	trimmed = strings.TrimRight(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, ".git")
	last := strings.LastIndexAny(trimmed, "/:")
	if last >= 0 {
		trimmed = trimmed[last+1:]
	}
	return normalizeTrackedPath(trimmed)
}

func validateTrackedRepo(file manifest.File, name, path string) error {
	for _, repo := range file.Repos {
		if repo.Name == name {
			return fmt.Errorf("repository name '%s' is already tracked", name)
		}
		if repo.Path == path {
			return fmt.Errorf("repository path '%s' is already tracked", path)
		}
		if pathOverlaps(repo.Path, path) {
			return fmt.Errorf("repository path '%s' overlaps tracked path '%s'", path, repo.Path)
		}
	}
	return nil
}

func pathOverlaps(left, right string) bool {
	left = strings.Trim(normalizeTrackedPath(left), "/")
	right = strings.Trim(normalizeTrackedPath(right), "/")
	if left == right {
		return true
	}
	return strings.HasPrefix(left, right+"/") || strings.HasPrefix(right, left+"/")
}

func normalizeTrackedPath(path string) string {
	return filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
}

func validateTrackedPath(path string) error {
	if path == "" || path == "." {
		return fmt.Errorf("destination path is required")
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("destination path must be relative: %s", path)
	}
	if path == ".." || strings.HasPrefix(path, "../") {
		return fmt.Errorf("destination path must stay within the workspace: %s", path)
	}
	return nil
}

func init() {
	trackCmd.Flags().StringVarP(&trackBranch, "branch", "b", "", "Branch of the repository to clone and record")
}
