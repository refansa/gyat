package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/refansa/gyat/internal/git"
	"github.com/refansa/gyat/internal/manifest"
	"github.com/refansa/gyat/internal/workspace"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a gyat workspace",
	Long: `Initialize a gyat workspace in the current directory.

This creates a .gyat manifest if one does not already exist and initializes
the root git repository if needed. Running the command again validates the
existing manifest and reinitializes git cleanly.

Nested gyat workspaces are not allowed: if the current directory is already
inside another gyat workspace, init fails instead of creating a second one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runInit(dir, cmd, args)
	},
}

func runInit(dir string, cmd *cobra.Command, args []string) error {
	root, err := workspace.FindRoot(dir)
	switch {
	case err == nil:
		same, sameErr := samePath(root, dir)
		if sameErr != nil {
			return sameErr
		}
		if !same {
			return fmt.Errorf("cannot initialize gyat workspace inside existing workspace rooted at %s", root)
		}
	case !errors.Is(err, workspace.ErrNotFound):
		return err
	}

	file := manifest.Default()
	hasManifest := false
	if _, err := os.Stat(manifest.FilePath(dir)); err == nil {
		hasManifest = true
		file, err = manifest.LoadDir(dir)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	out, err := git.Run(dir, "init")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.ErrOrStderr(), out)

	if hasManifest {
		changed, err := workspace.SyncGitIgnore(dir, file)
		if err != nil {
			return err
		}
		if changed {
			fmt.Fprintln(cmd.ErrOrStderr(), "updated .gitignore managed block")
		}
		fmt.Fprintln(cmd.ErrOrStderr(), "validated .gyat manifest")
		return nil
	}

	if err := manifest.SaveDir(dir, file); err != nil {
		return err
	}
	changed, err := workspace.SyncGitIgnore(dir, file)
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintln(cmd.ErrOrStderr(), "updated .gitignore managed block")
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "created .gyat manifest")
	fmt.Fprintln(cmd.ErrOrStderr(), "hint: use 'gyat track <repo> [path]' to start adding repositories")

	return nil
}

func samePath(left, right string) (bool, error) {
	leftPath, err := normalizePath(left)
	if err != nil {
		return false, err
	}
	rightPath, err := normalizePath(right)
	if err != nil {
		return false, err
	}

	return leftPath == rightPath, nil
}

func normalizePath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = resolved
	}

	return filepath.Clean(path), nil
}
