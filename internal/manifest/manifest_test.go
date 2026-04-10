package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	file := Default()

	if file.Version != SupportedVersion {
		t.Fatalf("Default version = %d, want %d", file.Version, SupportedVersion)
	}
	if file.Ignore == nil {
		t.Fatal("Default ignore should be initialized")
	}
	if file.Repos == nil {
		t.Fatal("Default repos should be initialized")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := File{
		Version: SupportedVersion,
		Ignore:  []string{" .vscode/ ", "node_modules/", ".vscode/"},
		Repos: []Repo{{
			Name:   " auth ",
			Path:   "services/auth/",
			URL:    " git@github.com:org/service-auth.git ",
			Branch: " main ",
			Groups: []string{"backend", " backend ", "api"},
		}},
	}

	if err := SaveDir(dir, input); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	got, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	if got.Version != SupportedVersion {
		t.Fatalf("Version = %d, want %d", got.Version, SupportedVersion)
	}
	if len(got.Ignore) != 2 || got.Ignore[0] != ".vscode/" || got.Ignore[1] != "node_modules/" {
		t.Fatalf("Ignore = %#v, want normalized unique values", got.Ignore)
	}
	if len(got.Repos) != 1 {
		t.Fatalf("Repos len = %d, want 1", len(got.Repos))
	}
	repo := got.Repos[0]
	if repo.Name != "auth" {
		t.Fatalf("Name = %q, want %q", repo.Name, "auth")
	}
	if repo.Path != "services/auth" {
		t.Fatalf("Path = %q, want %q", repo.Path, "services/auth")
	}
	if repo.URL != "git@github.com:org/service-auth.git" {
		t.Fatalf("URL = %q", repo.URL)
	}
	if repo.Branch != "main" {
		t.Fatalf("Branch = %q, want %q", repo.Branch, "main")
	}
	if len(repo.Groups) != 2 || repo.Groups[0] != "backend" || repo.Groups[1] != "api" {
		t.Fatalf("Groups = %#v, want normalized unique values", repo.Groups)
	}
}

func TestSaveDefaultsVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := SaveDir(dir, File{}); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	got, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if got.Version != SupportedVersion {
		t.Fatalf("Version = %d, want %d", got.Version, SupportedVersion)
	}
}

func TestValidateRejectsDuplicateRepoPaths(t *testing.T) {
	t.Parallel()

	err := SaveDir(t.TempDir(), File{
		Version: SupportedVersion,
		Ignore:  []string{},
		Repos: []Repo{
			{Name: "auth", Path: "services/auth", URL: "git@github.com:org/auth.git"},
			{Name: "billing", Path: "services/auth", URL: "git@github.com:org/billing.git"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate repo path") {
		t.Fatalf("SaveDir error = %v, want duplicate repo path", err)
	}
}

func TestLoadRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := FilePath(dir)
	if err := os.WriteFile(path, []byte("{\n  \"version\": 2,\n  \"ignore\": [],\n  \"repos\": []\n}\n"), 0o644); err != nil {
		t.Fatalf("Write manifest: %v", err)
	}

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported manifest version") {
		t.Fatalf("Load error = %v, want unsupported manifest version", err)
	}
}

func TestValidateRejectsAbsoluteRepoPath(t *testing.T) {
	t.Parallel()

	absPath, err := filepath.Abs("service-auth")
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}

	err = SaveDir(t.TempDir(), File{
		Version: SupportedVersion,
		Ignore:  []string{},
		Repos: []Repo{{
			Name: "auth",
			Path: absPath,
			URL:  "git@github.com:org/auth.git",
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "must be relative") {
		t.Fatalf("SaveDir error = %v, want relative path validation", err)
	}
}
