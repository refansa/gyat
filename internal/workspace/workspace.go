package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/v2/internal/manifest"
)

// ErrNotFound indicates that no gyat workspace root could be found.
var ErrNotFound = errors.New("gyat workspace not found")

// Workspace is a loaded gyat workspace rooted at RootDir.
type Workspace struct {
	RootDir  string
	Manifest manifest.File
}

// FindRoot walks upward from start until it finds a .gyat manifest.
// If start is empty, the current working directory is used.
func FindRoot(start string) (string, error) {
	dir, err := normalizeStart(start)
	if err != nil {
		return "", err
	}

	for {
		manifestPath := manifest.FilePath(dir)
		info, statErr := os.Stat(manifestPath)
		if statErr == nil {
			if info.IsDir() {
				return "", fmt.Errorf("%s is a directory", manifest.FileName)
			}
			return dir, nil
		}
		if !os.IsNotExist(statErr) {
			return "", statErr
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// Load finds and loads a gyat workspace from start.
func Load(start string) (Workspace, error) {
	root, err := FindRoot(start)
	if err != nil {
		return Workspace{}, err
	}

	file, err := manifest.LoadDir(root)
	if err != nil {
		return Workspace{}, err
	}

	return Workspace{RootDir: root, Manifest: file}, nil
}

func normalizeStart(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get current directory: %w", err)
		}
	}

	start, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	info, err := os.Stat(start)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		start = filepath.Dir(start)
	}

	resolved, err := filepath.EvalSymlinks(start)
	if err == nil {
		start = resolved
	}

	return filepath.Clean(start), nil
}
