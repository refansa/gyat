package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/v2/internal/git"
	"github.com/refansa/gyat/v2/internal/manifest"
	"github.com/refansa/gyat/v2/internal/workspace"
)

// resolveWorkspaceRepoSelectors resolves target arguments to repo paths.
// It matches repo names and paths from the manifest, filtering out invalid args.
func resolveWorkspaceRepoSelectors(ws workspace.Workspace, startDir string, args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}

	repoPaths := make([]string, 0, len(args))
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		if trimmed == "" {
			continue
		}

		for _, repo := range ws.Manifest.Repos {
			if repo.Name == trimmed {
				repoPaths = append(repoPaths, repo.Path)
				break
			}
		}

		rel, err := normalizeWorkspaceArg(ws.RootDir, startDir, trimmed)
		if err != nil {
			continue
		}

		repoPath, _, _, matched := matchTrackedRepo(ws.Manifest.Repos, rel)
		if matched {
			repoPaths = append(repoPaths, repoPath)
		}
	}

	return repoPaths, nil
}

// workspaceRepoByPath finds a repo in the workspace manifest by its path.
func workspaceRepoByPath(ws workspace.Workspace, path string) (manifest.Repo, bool) {
	for _, repo := range ws.Manifest.Repos {
		if repo.Path == path {
			return repo, true
		}
	}
	return manifest.Repo{}, false
}

// hasUpstream checks if a directory has a remote configured.
func hasUpstream(dir string) bool {
	out, err := git.Run(dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	return err == nil && strings.TrimSpace(out) != ""
}

// normalizeWorkspaceArg resolves a path argument to be relative to the workspace root.
// Returns error if the path escapes the workspace.
func normalizeWorkspaceArg(root, startDir, arg string) (string, error) {
	if strings.TrimSpace(arg) == "" {
		return "", fmt.Errorf("path is required")
	}

	clean := strings.TrimLeft(arg, "/")
	clean = strings.TrimPrefix(clean, "./")

	absArg := arg
	if !filepath.IsAbs(clean) {
		absArg = filepath.Join(startDir, clean)
	}

	absRoot := filepath.Clean(root)
	absArg = filepath.Clean(absArg)

	if !strings.HasPrefix(absArg, absRoot) {
		return "", fmt.Errorf("path '%s' escapes workspace root", arg)
	}

	rel, err := filepath.Rel(absRoot, absArg)
	if err != nil {
		return "", err
	}

	return rel, nil
}

// matchTrackedRepo checks if the given argument matches a tracked repo path.
// Returns (repoPath, repoArg, stageAll, matched).
func matchTrackedRepo(repos []manifest.Repo, arg string) (repoPath, repoArg string, stageAll bool, matched bool) {
	if arg == "" {
		return "", "", false, false
	}

	argNorm := filepath.ToSlash(arg)

	for _, repo := range repos {
		repoPathNorm := filepath.ToSlash(repo.Path)

		if argNorm == repoPathNorm {
			return repo.Path, "", true, true
		}

		if strings.HasPrefix(argNorm, repoPathNorm+"/") {
			rest := strings.TrimPrefix(argNorm, repoPathNorm+"/")
			return repo.Path, rest, false, true
		}
	}

	return "", "", false, false
}