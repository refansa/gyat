package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/refansa/gyat/internal/manifest"
)

func TestSyncGitIgnoreCreatesManagedBlock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	changed, err := SyncGitIgnore(dir, manifest.File{
		Version: manifest.SupportedVersion,
		Ignore:  []string{".vscode/"},
		Repos: []manifest.Repo{{
			Name: "auth",
			Path: "services/auth",
			URL:  "git@github.com:org/auth.git",
		}},
	})
	if err != nil {
		t.Fatalf("SyncGitIgnore: %v", err)
	}
	if !changed {
		t.Fatal("SyncGitIgnore changed = false, want true")
	}

	data, err := os.ReadFile(filepath.Join(dir, gitIgnoreFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want := strings.Join([]string{
		managedBlockStartLine,
		".vscode/",
		"/services/auth/",
		managedBlockEndLine,
		"",
	}, "\n")
	if string(data) != want {
		t.Fatalf(".gitignore = %q, want %q", string(data), want)
	}
}

func TestSyncGitIgnorePreservesUserContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, gitIgnoreFileName)
	initial := strings.Join([]string{
		"node_modules/",
		".env",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	changed, err := SyncGitIgnore(dir, manifest.File{
		Version: manifest.SupportedVersion,
		Repos: []manifest.Repo{{
			Name: "auth",
			Path: "services/auth",
			URL:  "git@github.com:org/auth.git",
		}},
	})
	if err != nil {
		t.Fatalf("SyncGitIgnore: %v", err)
	}
	if !changed {
		t.Fatal("SyncGitIgnore changed = false, want true")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want := strings.Join([]string{
		"node_modules/",
		".env",
		"",
		managedBlockStartLine,
		"/services/auth/",
		managedBlockEndLine,
		"",
	}, "\n")
	if string(data) != want {
		t.Fatalf(".gitignore = %q, want %q", string(data), want)
	}
}

func TestSyncGitIgnoreUpdatesExistingBlock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, gitIgnoreFileName)
	initial := strings.Join([]string{
		".env",
		"",
		managedBlockStartLine,
		"/services/old/",
		managedBlockEndLine,
		"",
		"coverage.out",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	changed, err := SyncGitIgnore(dir, manifest.File{
		Version: manifest.SupportedVersion,
		Ignore:  []string{".vscode/"},
		Repos: []manifest.Repo{{
			Name: "auth",
			Path: "services/auth",
			URL:  "git@github.com:org/auth.git",
		}},
	})
	if err != nil {
		t.Fatalf("SyncGitIgnore: %v", err)
	}
	if !changed {
		t.Fatal("SyncGitIgnore changed = false, want true")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want := strings.Join([]string{
		".env",
		"",
		managedBlockStartLine,
		".vscode/",
		"/services/auth/",
		managedBlockEndLine,
		"",
		"coverage.out",
		"",
	}, "\n")
	if string(data) != want {
		t.Fatalf(".gitignore = %q, want %q", string(data), want)
	}
}

func TestSyncGitIgnoreRemovesManagedBlockWhenEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, gitIgnoreFileName)
	initial := strings.Join([]string{
		managedBlockStartLine,
		"/services/auth/",
		managedBlockEndLine,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	changed, err := SyncGitIgnore(dir, manifest.Default())
	if err != nil {
		t.Fatalf("SyncGitIgnore: %v", err)
	}
	if !changed {
		t.Fatal("SyncGitIgnore changed = false, want true")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Stat .gitignore error = %v, want file to be removed", err)
	}
}

func TestSyncGitIgnoreRejectsMalformedManagedBlock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, gitIgnoreFileName)
	initial := strings.Join([]string{
		managedBlockStartLine,
		"/services/auth/",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := SyncGitIgnore(dir, manifest.Default())
	if err == nil || !strings.Contains(err.Error(), "incomplete gyat managed block") {
		t.Fatalf("SyncGitIgnore error = %v, want malformed block error", err)
	}
}
