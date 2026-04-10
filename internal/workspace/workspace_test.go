package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/refansa/gyat/internal/manifest"
)

func TestFindRootWalksUpwards(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := manifest.SaveDir(root, manifest.Default()); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	nested := filepath.Join(root, "services", "auth")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := FindRoot(nested)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if got != root {
		t.Fatalf("FindRoot = %q, want %q", got, root)
	}
}

func TestFindRootAcceptsFilePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := manifest.SaveDir(root, manifest.Default()); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	nested := filepath.Join(root, "notes")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	filePath := filepath.Join(nested, "workspace.txt")
	if err := os.WriteFile(filePath, []byte("test\n"), 0o644); err != nil {
		t.Fatalf("Write file: %v", err)
	}

	got, err := FindRoot(filePath)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if got != root {
		t.Fatalf("FindRoot = %q, want %q", got, root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	t.Parallel()

	_, err := FindRoot(t.TempDir())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindRoot error = %v, want ErrNotFound", err)
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := manifest.File{
		Version: manifest.SupportedVersion,
		Ignore:  []string{"node_modules/"},
		Repos: []manifest.Repo{{
			Name:   "auth",
			Path:   "services/auth",
			URL:    "git@github.com:org/service-auth.git",
			Branch: "main",
		}},
	}
	if err := manifest.SaveDir(root, want); err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	got, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.RootDir != root {
		t.Fatalf("RootDir = %q, want %q", got.RootDir, root)
	}
	if got.Manifest.Version != want.Version {
		t.Fatalf("Version = %d, want %d", got.Manifest.Version, want.Version)
	}
	if len(got.Manifest.Repos) != 1 || got.Manifest.Repos[0].Name != "auth" {
		t.Fatalf("Manifest repos = %#v, want loaded manifest", got.Manifest.Repos)
	}
}
