package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/refansa/gyat/internal/manifest"
)

// TargetOptions controls which workspace targets are selected.
type TargetOptions struct {
	IncludeRoot   bool
	RootOnly      bool
	RepoSelectors []string
	Groups        []string
}

// Target is one execution target within a gyat workspace.
type Target struct {
	Label  string
	Dir    string
	Name   string
	Path   string
	Groups []string
	IsRoot bool
}

// ResolveTargets selects the root workspace and/or managed repos according to options.
func (workspace Workspace) ResolveTargets(options TargetOptions) ([]Target, error) {
	if options.RootOnly {
		if !options.IncludeRoot {
			return nil, fmt.Errorf("root-only cannot be combined with no-root")
		}
		if len(options.RepoSelectors) > 0 || len(options.Groups) > 0 {
			return nil, fmt.Errorf("root-only cannot be combined with repo or group selectors")
		}
		return []Target{rootTarget(workspace.RootDir)}, nil
	}

	selectors := normalizedUnique(options.RepoSelectors, normalizeSelector)
	groups := normalizedUnique(options.Groups, normalizeGroup)
	includeAllRepos := len(selectors) == 0 && len(groups) == 0

	matchedSelectors := make(map[string]struct{}, len(selectors))
	matchedGroups := make(map[string]struct{}, len(groups))
	selected := make([]Target, 0, len(workspace.Manifest.Repos)+1)

	if options.IncludeRoot {
		selected = append(selected, rootTarget(workspace.RootDir))
	}

	for _, repo := range workspace.Manifest.Repos {
		if !includeAllRepos && !repoSelected(repo, selectors, groups, matchedSelectors, matchedGroups) {
			continue
		}
		selected = append(selected, repoTarget(workspace.RootDir, repo))
	}

	if len(selectors) > len(matchedSelectors) {
		return nil, fmt.Errorf("unknown repo selector(s): %s", joinMissing(selectors, matchedSelectors))
	}
	if len(groups) > len(matchedGroups) {
		return nil, fmt.Errorf("unknown group selector(s): %s", joinMissing(groups, matchedGroups))
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no targets selected")
	}

	return selected, nil
}

func rootTarget(root string) Target {
	return Target{
		Label:  "umbrella repository",
		Dir:    root,
		Name:   ".",
		Path:   ".",
		IsRoot: true,
	}
}

func repoTarget(root string, repo manifest.Repo) Target {
	return Target{
		Label:  repo.Path,
		Dir:    filepath.Join(root, filepath.FromSlash(repo.Path)),
		Name:   repo.Name,
		Path:   repo.Path,
		Groups: append([]string(nil), repo.Groups...),
	}
}

func repoSelected(repo manifest.Repo, selectors, groups []string, matchedSelectors, matchedGroups map[string]struct{}) bool {
	selected := false
	for _, selector := range selectors {
		if selectorMatchesRepo(repo, selector) {
			matchedSelectors[selector] = struct{}{}
			selected = true
		}
	}
	for _, group := range groups {
		if repoInGroup(repo, group) {
			matchedGroups[group] = struct{}{}
			selected = true
		}
	}
	return selected
}

func selectorMatchesRepo(repo manifest.Repo, selector string) bool {
	return repo.Name == selector || normalizeSelector(repo.Path) == selector
}

func repoInGroup(repo manifest.Repo, group string) bool {
	for _, candidate := range repo.Groups {
		if normalizeGroup(candidate) == group {
			return true
		}
	}
	return false
}

func normalizeSelector(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func normalizeGroup(value string) string {
	return strings.TrimSpace(value)
}

func normalizedUnique(values []string, normalize func(string) string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = normalize(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func joinMissing(expected []string, seen map[string]struct{}) string {
	missing := make([]string, 0, len(expected))
	for _, value := range expected {
		if _, ok := seen[value]; ok {
			continue
		}
		missing = append(missing, value)
	}
	return strings.Join(missing, ", ")
}
