package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// execDir returns the directory gyat should treat as the repository root.
//
// When the binary is invoked by path (e.g. ../b/gyat), it returns the
// directory that contains the binary so that gyat always targets the
// repository it lives in, regardless of the caller's working directory.
//
// When the binary was resolved through PATH (e.g. it is installed globally in
// ~/go/bin or /usr/local/bin), that directory is not a git repository, so we
// fall back to the caller's current working directory instead — matching the
// behaviour of plain git.
func execDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}
	exeDir := filepath.Clean(filepath.Dir(exe))

	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == "" {
			continue
		}
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			resolved = p
		}
		if exeDir == filepath.Clean(resolved) {
			// Binary lives in a PATH directory — fall back to cwd.
			return os.Getwd()
		}
	}

	return exeDir, nil
}
