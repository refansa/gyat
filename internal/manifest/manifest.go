package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// FileName is the workspace manifest stored at the umbrella root.
	FileName = ".gyat"

	// SupportedVersion is the only manifest schema version currently supported.
	SupportedVersion = 1
)

// File is gyat's workspace manifest.
type File struct {
	Version int      `json:"version"`
	Ignore  []string `json:"ignore"`
	Repos   []Repo   `json:"repos"`
}

// Repo is one managed repository within a gyat workspace.
type Repo struct {
	Name   string   `json:"name"`
	Path   string   `json:"path"`
	URL    string   `json:"url"`
	Branch string   `json:"branch,omitempty"`
	Groups []string `json:"groups,omitempty"`
}

// Default returns an empty manifest for a new workspace.
func Default() File {
	return File{
		Version: SupportedVersion,
		Ignore:  []string{},
		Repos:   []Repo{},
	}
}

// FilePath returns the manifest path for a workspace directory.
func FilePath(dir string) string {
	return filepath.Join(dir, FileName)
}

// LoadDir loads and validates the manifest in dir.
func LoadDir(dir string) (File, error) {
	return Load(FilePath(dir))
}

// SaveDir validates and writes the manifest in dir.
func SaveDir(dir string, file File) error {
	return Save(FilePath(dir), file)
}

// Load reads a manifest file from disk.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	if strings.TrimSpace(string(data)) == "" {
		return File{}, fmt.Errorf("manifest %s is empty", filepath.Base(path))
	}

	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return File{}, fmt.Errorf("parse manifest %s: %w", filepath.Base(path), err)
	}

	file = normalizeFile(file)
	if err := Validate(file); err != nil {
		return File{}, err
	}

	return file, nil
}

// Save validates and writes a manifest file to disk.
func Save(path string, file File) error {
	if file.Version == 0 {
		file.Version = SupportedVersion
	}
	file = normalizeFile(file)

	if err := Validate(file); err != nil {
		return err
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest %s: %w", filepath.Base(path), err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write manifest %s: %w", filepath.Base(path), err)
	}

	return nil
}

// Validate ensures a manifest can be safely used by gyat.
func Validate(file File) error {
	if file.Version == 0 {
		return fmt.Errorf("manifest version is required")
	}
	if file.Version != SupportedVersion {
		return fmt.Errorf("unsupported manifest version %d", file.Version)
	}

	names := map[string]struct{}{}
	paths := map[string]struct{}{}

	for index, repo := range file.Repos {
		if repo.Name == "" {
			return fmt.Errorf("repos[%d].name is required", index)
		}
		if repo.Path == "" || repo.Path == "." {
			return fmt.Errorf("repos[%d].path is required", index)
		}
		if filepath.IsAbs(repo.Path) {
			return fmt.Errorf("repos[%d].path must be relative: %s", index, repo.Path)
		}
		if repo.Path == ".." || strings.HasPrefix(repo.Path, "../") {
			return fmt.Errorf("repos[%d].path must stay within the workspace: %s", index, repo.Path)
		}
		if repo.URL == "" {
			return fmt.Errorf("repos[%d].url is required", index)
		}

		if _, exists := names[repo.Name]; exists {
			return fmt.Errorf("duplicate repo name %q", repo.Name)
		}
		if _, exists := paths[repo.Path]; exists {
			return fmt.Errorf("duplicate repo path %q", repo.Path)
		}

		names[repo.Name] = struct{}{}
		paths[repo.Path] = struct{}{}
	}

	return nil
}

func normalizeFile(file File) File {
	file.Ignore = normalizeStrings(file.Ignore)
	if file.Ignore == nil {
		file.Ignore = []string{}
	}

	if file.Repos == nil {
		file.Repos = []Repo{}
	}
	for index, repo := range file.Repos {
		file.Repos[index] = normalizeRepo(repo)
	}

	return file
}

func normalizeRepo(repo Repo) Repo {
	repo.Name = strings.TrimSpace(repo.Name)
	repo.Path = normalizeRepoPath(repo.Path)
	repo.URL = strings.TrimSpace(repo.URL)
	repo.Branch = strings.TrimSpace(repo.Branch)
	repo.Groups = normalizeStrings(repo.Groups)

	return repo
}

func normalizeRepoPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	return filepath.ToSlash(filepath.Clean(path))
}

func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		value = strings.ReplaceAll(value, `\`, "/")
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}
