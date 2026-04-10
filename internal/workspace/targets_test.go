package workspace

import (
	"strings"
	"testing"

	"github.com/refansa/gyat/internal/manifest"
)

func TestResolveTargetsIncludesRootAndReposByDefault(t *testing.T) {
	t.Parallel()

	workspace := Workspace{
		RootDir: t.TempDir(),
		Manifest: manifest.File{
			Version: manifest.SupportedVersion,
			Repos: []manifest.Repo{
				{Name: "auth", Path: "services/auth", URL: "git@github.com:org/auth.git", Groups: []string{"backend"}},
				{Name: "billing", Path: "services/billing", URL: "git@github.com:org/billing.git", Groups: []string{"payments"}},
			},
		},
	}

	targets, err := workspace.ResolveTargets(TargetOptions{IncludeRoot: true})
	if err != nil {
		t.Fatalf("ResolveTargets: %v", err)
	}
	if len(targets) != 3 {
		t.Fatalf("ResolveTargets len = %d, want 3", len(targets))
	}
	if !targets[0].IsRoot || targets[0].Label != "umbrella repository" {
		t.Fatalf("targets[0] = %#v, want root target", targets[0])
	}
	if targets[1].Path != "services/auth" || targets[2].Path != "services/billing" {
		t.Fatalf("repo targets = %#v, want manifest order", targets[1:])
	}
}

func TestResolveTargetsFiltersByRepoAndGroup(t *testing.T) {
	t.Parallel()

	workspace := Workspace{
		RootDir: t.TempDir(),
		Manifest: manifest.File{
			Version: manifest.SupportedVersion,
			Repos: []manifest.Repo{
				{Name: "auth", Path: "services/auth", URL: "git@github.com:org/auth.git", Groups: []string{"backend", "api"}},
				{Name: "billing", Path: "services/billing", URL: "git@github.com:org/billing.git", Groups: []string{"payments"}},
			},
		},
	}

	targets, err := workspace.ResolveTargets(TargetOptions{
		IncludeRoot:   false,
		RepoSelectors: []string{"services/auth"},
		Groups:        []string{"payments"},
	})
	if err != nil {
		t.Fatalf("ResolveTargets: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("ResolveTargets len = %d, want 2", len(targets))
	}
	if targets[0].Name != "auth" || targets[1].Name != "billing" {
		t.Fatalf("targets = %#v, want auth and billing", targets)
	}
}

func TestResolveTargetsRejectsUnknownSelectors(t *testing.T) {
	t.Parallel()

	workspace := Workspace{RootDir: t.TempDir(), Manifest: manifest.Default()}
	_, err := workspace.ResolveTargets(TargetOptions{RepoSelectors: []string{"ghost"}})
	if err == nil || !strings.Contains(err.Error(), "unknown repo selector") {
		t.Fatalf("ResolveTargets error = %v, want unknown repo selector", err)
	}
}

func TestResolveTargetsRejectsInvalidRootOnlyCombination(t *testing.T) {
	t.Parallel()

	workspace := Workspace{RootDir: t.TempDir(), Manifest: manifest.Default()}
	_, err := workspace.ResolveTargets(TargetOptions{IncludeRoot: true, RootOnly: true, RepoSelectors: []string{"auth"}})
	if err == nil || !strings.Contains(err.Error(), "root-only") {
		t.Fatalf("ResolveTargets error = %v, want root-only validation", err)
	}
}
