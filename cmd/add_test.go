package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Helpers specific to staging tests
// ---------------------------------------------------------------------------

// setupTrackedSubmodule creates an umbrella repo with a single tracked
// submodule and returns the umbrella directory and the submodule path within
// it. The source repo is named subName and is placed as a sibling of the
// umbrella directory so that the relative path ../subName is valid.
func setupTrackedSubmodule(t *testing.T, subName string) (umbrella, subPath string) {
	t.Helper()
	skipIfNoGit(t)

	umbrella, source := newTestSetup(t, subName)
	rel := relPath(umbrella, source)

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{rel}); err != nil {
		t.Fatalf("setupTrackedSubmodule: runTrack %q: %v", rel, err)
	}

	return umbrella, subName
}

// dirtySubmodule writes a new file into subDir, leaving the submodule working
// tree with an untracked file that is ready to be staged.
func dirtySubmodule(t *testing.T, subDir string) {
	t.Helper()
	writeFile(t, filepath.Join(subDir, "new_feature.go"), "package main\n// new feature\n")
}

// stagedFilesInDir returns the newline-separated list of files that are
// currently staged (in the index) inside dir.
func stagedFilesInDir(t *testing.T, dir string) string {
	t.Helper()
	return runGitIn(t, dir, "diff", "--cached", "--name-only")
}

// ---------------------------------------------------------------------------
// Unit tests — hasWorkingTreeChanges
// ---------------------------------------------------------------------------

// TestHasWorkingTreeChanges_Empty verifies that an empty status output is
// treated as having no working-tree changes.
func TestHasWorkingTreeChanges_Empty(t *testing.T) {
	t.Parallel()
	if hasWorkingTreeChanges("") {
		t.Error("expected false for empty status output")
	}
}

// TestHasWorkingTreeChanges_StagedOnly verifies that lines where Y=' '
// (index-only change, working tree clean) do not trigger a true result.
func TestHasWorkingTreeChanges_StagedOnly(t *testing.T) {
	t.Parallel()
	// "A  file" → X='A', Y=' ': staged new file, working tree matches index.
	out := "A  .gitmodules\nA  services/auth\n"
	if hasWorkingTreeChanges(out) {
		t.Errorf("expected false for staged-only output\ninput: %q", out)
	}
}

// TestHasWorkingTreeChanges_Untracked verifies that an untracked file (Y='?')
// is detected as a working-tree change.
func TestHasWorkingTreeChanges_Untracked(t *testing.T) {
	t.Parallel()
	out := "?? new_file.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for untracked file\ninput: %q", out)
	}
}

// TestHasWorkingTreeChanges_ModifiedInWorkingTree verifies that a file
// modified in the working tree (Y='M') is detected.
func TestHasWorkingTreeChanges_ModifiedInWorkingTree(t *testing.T) {
	t.Parallel()
	out := " M main.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for working-tree modification\ninput: %q", out)
	}
}

// TestHasWorkingTreeChanges_DeletedInWorkingTree verifies that a file
// deleted in the working tree (Y='D') is detected.
func TestHasWorkingTreeChanges_DeletedInWorkingTree(t *testing.T) {
	t.Parallel()
	out := " D old_file.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for working-tree deletion\ninput: %q", out)
	}
}

// TestHasWorkingTreeChanges_StagedAndModified verifies that a file staged as
// new but also modified again in the working tree (e.g. "AM") is detected.
func TestHasWorkingTreeChanges_StagedAndModified(t *testing.T) {
	t.Parallel()
	out := "AM handler.go\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true for staged-and-modified file\ninput: %q", out)
	}
}

// TestHasWorkingTreeChanges_MixedStagedAndUntracked verifies that a mix of
// staged-only and untracked lines correctly returns true.
func TestHasWorkingTreeChanges_MixedStagedAndUntracked(t *testing.T) {
	t.Parallel()
	out := "A  .gitmodules\n?? scratch.txt\n"
	if !hasWorkingTreeChanges(out) {
		t.Errorf("expected true when output includes an untracked file\ninput: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — classifyArgs
// ---------------------------------------------------------------------------

// TestClassifyArgs_NoArgs verifies that empty input produces empty outputs.
func TestClassifyArgs_NoArgs(t *testing.T) {
	t.Parallel()
	rootArgs, subTargets := classifyArgs([]string{"services/auth"}, nil)
	if len(rootArgs) != 0 {
		t.Errorf("rootArgs = %v, want empty", rootArgs)
	}
	if len(subTargets) != 0 {
		t.Errorf("subTargets = %v, want empty", subTargets)
	}
}

// TestClassifyArgs_RootFile verifies that a path outside any submodule is
// placed in rootArgs.
func TestClassifyArgs_RootFile(t *testing.T) {
	t.Parallel()
	rootArgs, subTargets := classifyArgs([]string{"services/auth"}, []string{"README.md"})
	if len(rootArgs) != 1 || rootArgs[0] != "README.md" {
		t.Errorf("rootArgs = %v, want [README.md]", rootArgs)
	}
	if len(subTargets) != 0 {
		t.Errorf("subTargets = %v, want empty", subTargets)
	}
}

// TestClassifyArgs_ExactSubmoduleMatch verifies that a path equal to a
// submodule root sets stageAll=true.
func TestClassifyArgs_ExactSubmoduleMatch(t *testing.T) {
	t.Parallel()
	_, subTargets := classifyArgs([]string{"services/auth"}, []string{"services/auth"})
	stage, ok := subTargets["services/auth"]
	if !ok {
		t.Fatal("expected entry for services/auth in subTargets")
	}
	if !stage.stageAll {
		t.Error("expected stageAll=true for exact submodule match")
	}
}

// TestClassifyArgs_FileInsideSubmodule verifies that a path inside a submodule
// directory is routed as a specific file (stageAll=false).
func TestClassifyArgs_FileInsideSubmodule(t *testing.T) {
	t.Parallel()
	_, subTargets := classifyArgs([]string{"services/auth"}, []string{"services/auth/handler.go"})
	stage, ok := subTargets["services/auth"]
	if !ok {
		t.Fatal("expected entry for services/auth in subTargets")
	}
	if stage.stageAll {
		t.Error("expected stageAll=false for file-inside-submodule")
	}
	if len(stage.files) != 1 || stage.files[0] != "handler.go" {
		t.Errorf("files = %v, want [handler.go]", stage.files)
	}
}

// TestClassifyArgs_MultipleFilesInsideSameSubmodule verifies that multiple
// file paths inside the same submodule are collected together.
func TestClassifyArgs_MultipleFilesInsideSameSubmodule(t *testing.T) {
	t.Parallel()
	args := []string{"services/auth/handler.go", "services/auth/config.go"}
	_, subTargets := classifyArgs([]string{"services/auth"}, args)
	stage, ok := subTargets["services/auth"]
	if !ok {
		t.Fatal("expected entry for services/auth")
	}
	if stage.stageAll {
		t.Error("expected stageAll=false")
	}
	if len(stage.files) != 2 {
		t.Fatalf("files = %v, want 2 entries", stage.files)
	}
	fileSet := map[string]bool{stage.files[0]: true, stage.files[1]: true}
	for _, want := range []string{"handler.go", "config.go"} {
		if !fileSet[want] {
			t.Errorf("files missing %q, got %v", want, stage.files)
		}
	}
}

// TestClassifyArgs_ExactMatchSupersedesFiles verifies that once a submodule is
// targeted by its root path (stageAll=true), subsequent file paths inside the
// same submodule are ignored.
func TestClassifyArgs_ExactMatchSupersedesFiles(t *testing.T) {
	t.Parallel()
	// File first, then the submodule root — stageAll should win.
	args := []string{"services/auth/handler.go", "services/auth"}
	_, subTargets := classifyArgs([]string{"services/auth"}, args)
	stage, ok := subTargets["services/auth"]
	if !ok {
		t.Fatal("expected entry for services/auth")
	}
	if !stage.stageAll {
		t.Error("expected stageAll=true when submodule root is explicitly listed")
	}
}

// TestClassifyArgs_MixedRootAndSubmodule verifies that a mixed argument list
// is correctly split between rootArgs and per-submodule targets.
func TestClassifyArgs_MixedRootAndSubmodule(t *testing.T) {
	t.Parallel()
	subs := []string{"services/auth", "services/billing"}
	args := []string{"README.md", "services/auth", "services/billing/main.go"}
	rootArgs, subTargets := classifyArgs(subs, args)

	if len(rootArgs) != 1 || rootArgs[0] != "README.md" {
		t.Errorf("rootArgs = %v, want [README.md]", rootArgs)
	}

	authStage, ok := subTargets["services/auth"]
	if !ok {
		t.Fatal("expected entry for services/auth")
	}
	if !authStage.stageAll {
		t.Error("expected services/auth stageAll=true")
	}

	billingStage, ok := subTargets["services/billing"]
	if !ok {
		t.Fatal("expected entry for services/billing")
	}
	if billingStage.stageAll {
		t.Error("expected services/billing stageAll=false")
	}
	if len(billingStage.files) != 1 || billingStage.files[0] != "main.go" {
		t.Errorf("services/billing files = %v, want [main.go]", billingStage.files)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — allSubmodulePaths
// ---------------------------------------------------------------------------

// TestAllSubmodulePaths_NoGitmodules verifies that a repo with no .gitmodules
// returns an empty slice without error.
func TestAllSubmodulePaths_NoGitmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	paths, err := allSubmodulePaths(dir)
	if err != nil {
		t.Fatalf("allSubmodulePaths on repo with no .gitmodules: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected no paths, got %v", paths)
	}
}

// TestAllSubmodulePaths_ReturnsPaths verifies that a repo with a tracked
// submodule returns that submodule's path.
func TestAllSubmodulePaths_ReturnsPaths(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-paths")

	paths, err := allSubmodulePaths(umbrella)
	if err != nil {
		t.Fatalf("allSubmodulePaths: %v", err)
	}
	if len(paths) != 1 || paths[0] != subPath {
		t.Errorf("allSubmodulePaths = %v, want [%s]", paths, subPath)
	}
}

// TestAllSubmodulePaths_ReturnsMultiplePaths verifies that all registered
// submodule paths are returned when there is more than one.
func TestAllSubmodulePaths_ReturnsMultiplePaths(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-a")
	svcB := filepath.Join(base, "svc-b")

	for _, d := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, d := range []string{svcA, svcB} {
		runGitIn(t, d, "init")
		runGitIn(t, d, "config", "user.email", "test@gyat.test")
		runGitIn(t, d, "config", "user.name", "gyat test")
		runGitIn(t, d, "config", "commit.gpgsign", "false")
		runGitIn(t, d, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(d, "main.go"), "package main\n")
		runGitIn(t, d, "add", ".")
		runGitIn(t, d, "commit", "-m", "init")
	}

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{"../svc-a"}); err != nil {
		t.Fatalf("track svc-a: %v", err)
	}
	if err := runTrack(umbrella, "", tc, []string{"../svc-b"}); err != nil {
		t.Fatalf("track svc-b: %v", err)
	}

	paths, err := allSubmodulePaths(umbrella)
	if err != nil {
		t.Fatalf("allSubmodulePaths: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %v", paths)
	}

	pathSet := map[string]bool{}
	for _, p := range paths {
		pathSet[p] = true
	}
	for _, want := range []string{"svc-a", "svc-b"} {
		if !pathSet[want] {
			t.Errorf("allSubmodulePaths missing %q, got %v", want, paths)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests — runAdd (no args)
// ---------------------------------------------------------------------------

// TestRunAdd_NoChanges_PrintsNothingToStage verifies that when neither the
// umbrella root nor any submodule has working-tree changes, the command
// prints "nothing to stage" and returns without error.
func TestRunAdd_NoChanges_PrintsNothingToStage(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(dir, cmd, nil); err != nil {
		t.Fatalf("runAdd on empty repo: %v", err)
	}
	if !strings.Contains(stderrBuf.String(), "nothing to stage") {
		t.Errorf("expected 'nothing to stage'\ngot: %q", stderrBuf.String())
	}
}

// TestRunAdd_StagesUmbrellaRoot verifies that an untracked file in the
// umbrella root is staged when no arguments are provided, even when there
// are no registered submodules.
func TestRunAdd_StagesUmbrellaRoot(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	dir := newUmbrellaRepo(t)
	// An initial commit gives git diff --cached a HEAD to compare against.
	writeFile(t, filepath.Join(dir, "README.md"), "# test\n")
	runGitIn(t, dir, "add", ".")
	runGitIn(t, dir, "commit", "-m", "initial commit")

	// Write an untracked file to the umbrella root.
	writeFile(t, filepath.Join(dir, ".gitignore"), "*.log\n")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(dir, cmd, nil); err != nil {
		t.Fatalf("runAdd: %v", err)
	}

	if staged := stagedFilesInDir(t, dir); !strings.Contains(staged, ".gitignore") {
		t.Errorf("expected .gitignore to be staged in umbrella root\ngit diff --cached:\n%s", staged)
	}
}

// TestRunAdd_NothingToStage_PrintsMessage verifies that when all submodule
// working trees are clean (only index-only changes present), a
// 'nothing to stage' message is printed.
func TestRunAdd_NothingToStage_PrintsMessage(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	// setupTrackedSubmodule leaves the umbrella with staged-only changes
	// (.gitmodules, submodule pointer) and the submodule with a clean working tree.
	umbrella, _ := setupTrackedSubmodule(t, "svc-clean")

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd on clean submodules: %v", err)
	}
	if !strings.Contains(stderrBuf.String(), "nothing to stage") {
		t.Errorf("expected 'nothing to stage' message\ngot: %q", stderrBuf.String())
	}
}

// TestRunAdd_StagesAllSubmodules verifies that runAdd with no args stages
// pending changes across all registered submodules.
func TestRunAdd_StagesAllSubmodules(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-stage-all")
	subDir := filepath.Join(umbrella, subPath)

	dirtySubmodule(t, subDir)

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd: %v", err)
	}

	if staged := stagedFilesInDir(t, subDir); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected 'new_feature.go' to be staged in %s\ngit diff --cached:\n%s", subPath, staged)
	}
}

// TestRunAdd_StagesRootAndSubmodule verifies that a single no-arg invocation
// stages changes in both the umbrella root and a dirty submodule.
func TestRunAdd_StagesRootAndSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-root-and-sub")
	subDir := filepath.Join(umbrella, subPath)

	// Commit the tracked submodule so the umbrella has a clean HEAD to diff against.
	runGitIn(t, umbrella, "commit", "-m", "track submodule")

	// Dirty both the umbrella root and the submodule.
	writeFile(t, filepath.Join(umbrella, ".gitignore"), "*.log\n")
	dirtySubmodule(t, subDir)

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd: %v", err)
	}

	if staged := stagedFilesInDir(t, umbrella); !strings.Contains(staged, ".gitignore") {
		t.Errorf("expected .gitignore to be staged in umbrella root\ngit diff --cached:\n%s", staged)
	}
	if staged := stagedFilesInDir(t, subDir); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected new_feature.go to be staged in submodule\ngit diff --cached:\n%s", staged)
	}
}

// TestRunAdd_UncheckedOutSubmodule_Warns verifies that a submodule whose
// directory is absent on disk (e.g. not yet initialized) is skipped with a
// warning rather than causing a fatal error.
func TestRunAdd_UncheckedOutSubmodule_Warns(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-absent")

	// Remove the checked-out submodule directory to simulate an
	// uninitialized / not-yet-cloned submodule.
	if err := os.RemoveAll(filepath.Join(umbrella, subPath)); err != nil {
		t.Fatalf("remove submodule dir: %v", err)
	}

	var stderrBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd with absent submodule dir should not error: %v", err)
	}
	if !strings.Contains(stderrBuf.String(), "warning") {
		t.Errorf("expected a warning about the absent submodule\ngot: %q", stderrBuf.String())
	}
}

// TestRunAdd_MultipleSubmodules_StagesAll verifies that all dirty submodules
// are staged when no path arguments are given.
func TestRunAdd_MultipleSubmodules_StagesAll(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-multi-a")
	svcB := filepath.Join(base, "svc-multi-b")

	for _, d := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, d := range []string{svcA, svcB} {
		runGitIn(t, d, "init")
		runGitIn(t, d, "config", "user.email", "test@gyat.test")
		runGitIn(t, d, "config", "user.name", "gyat test")
		runGitIn(t, d, "config", "commit.gpgsign", "false")
		runGitIn(t, d, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(d, "main.go"), "package main\n")
		runGitIn(t, d, "add", ".")
		runGitIn(t, d, "commit", "-m", "init")
	}

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{"../svc-multi-a"}); err != nil {
		t.Fatalf("track svc-multi-a: %v", err)
	}
	if err := runTrack(umbrella, "", tc, []string{"../svc-multi-b"}); err != nil {
		t.Fatalf("track svc-multi-b: %v", err)
	}

	// Dirty both submodules.
	dirtySubmodule(t, filepath.Join(umbrella, "svc-multi-a"))
	dirtySubmodule(t, filepath.Join(umbrella, "svc-multi-b"))

	// Stage everything with no args.
	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	if err := runAdd(umbrella, cmd, nil); err != nil {
		t.Fatalf("runAdd (no args): %v", err)
	}

	for _, name := range []string{"svc-multi-a", "svc-multi-b"} {
		if staged := stagedFilesInDir(t, filepath.Join(umbrella, name)); !strings.Contains(staged, "new_feature.go") {
			t.Errorf("expected %s to have staged changes\ngot: %q", name, staged)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests — runAdd (with args)
// ---------------------------------------------------------------------------

// TestRunAdd_StagesSpecificSubmodule verifies that when a submodule path is
// explicitly provided, only that submodule is staged and others are left
// untouched.
func TestRunAdd_StagesSpecificSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	base := t.TempDir()
	umbrella := filepath.Join(base, "umbrella")
	svcA := filepath.Join(base, "svc-a")
	svcB := filepath.Join(base, "svc-b")

	for _, d := range []string{umbrella, svcA, svcB} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	runGitIn(t, umbrella, "init")
	runGitIn(t, umbrella, "config", "user.email", "test@gyat.test")
	runGitIn(t, umbrella, "config", "user.name", "gyat test")
	runGitIn(t, umbrella, "config", "commit.gpgsign", "false")
	runGitIn(t, umbrella, "config", "core.autocrlf", "false")

	for _, d := range []string{svcA, svcB} {
		runGitIn(t, d, "init")
		runGitIn(t, d, "config", "user.email", "test@gyat.test")
		runGitIn(t, d, "config", "user.name", "gyat test")
		runGitIn(t, d, "config", "commit.gpgsign", "false")
		runGitIn(t, d, "config", "core.autocrlf", "false")
		writeFile(t, filepath.Join(d, "main.go"), "package main\n")
		runGitIn(t, d, "add", ".")
		runGitIn(t, d, "commit", "-m", "init")
	}

	tc := &cobra.Command{}
	tc.SetErr(io.Discard)
	if err := runTrack(umbrella, "", tc, []string{"../svc-a"}); err != nil {
		t.Fatalf("track svc-a: %v", err)
	}
	if err := runTrack(umbrella, "", tc, []string{"../svc-b"}); err != nil {
		t.Fatalf("track svc-b: %v", err)
	}

	// Dirty both submodules.
	dirtySubmodule(t, filepath.Join(umbrella, "svc-a"))
	dirtySubmodule(t, filepath.Join(umbrella, "svc-b"))

	// Stage only svc-a.
	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	if err := runAdd(umbrella, cmd, []string{"svc-a"}); err != nil {
		t.Fatalf("runAdd svc-a: %v", err)
	}

	// svc-a must have staged changes.
	if staged := stagedFilesInDir(t, filepath.Join(umbrella, "svc-a")); !strings.Contains(staged, "new_feature.go") {
		t.Errorf("expected svc-a to have staged changes\ngot: %q", staged)
	}
	// svc-b must NOT have staged changes.
	if staged := stagedFilesInDir(t, filepath.Join(umbrella, "svc-b")); staged != "" {
		t.Errorf("expected svc-b to have no staged changes, got: %q", staged)
	}
}

// TestRunAdd_ExplicitRootFile verifies that an explicit umbrella-root path is
// staged in the umbrella only, leaving submodule working trees untouched.
func TestRunAdd_ExplicitRootFile(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-explicit-root")
	subDir := filepath.Join(umbrella, subPath)

	// Commit the tracked submodule so the umbrella has a clean HEAD.
	runGitIn(t, umbrella, "commit", "-m", "track submodule")

	// Dirty both the umbrella root and the submodule.
	writeFile(t, filepath.Join(umbrella, ".gitignore"), "*.log\n")
	dirtySubmodule(t, subDir)

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	// Stage only the root .gitignore.
	if err := runAdd(umbrella, cmd, []string{".gitignore"}); err != nil {
		t.Fatalf("runAdd .gitignore: %v", err)
	}

	if staged := stagedFilesInDir(t, umbrella); !strings.Contains(staged, ".gitignore") {
		t.Errorf("expected .gitignore to be staged in umbrella root\ngit diff --cached:\n%s", staged)
	}
	if staged := stagedFilesInDir(t, subDir); staged != "" {
		t.Errorf("expected submodule to be untouched, got staged: %q", staged)
	}
}

// TestRunAdd_FileInsideSubmodule verifies that a path in the form
// "<submodule>/<file>" stages only the named file inside the submodule,
// leaving other files in the same submodule unstaged.
func TestRunAdd_FileInsideSubmodule(t *testing.T) {
	t.Parallel()
	skipIfNoGit(t)

	umbrella, subPath := setupTrackedSubmodule(t, "svc-file-route")
	subDir := filepath.Join(umbrella, subPath)

	// Write two untracked files inside the submodule — only one should be staged.
	writeFile(t, filepath.Join(subDir, "handler.go"), "package main\n")
	writeFile(t, filepath.Join(subDir, "other.go"), "package main\n")

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)

	// Stage only handler.go via its umbrella-relative path.
	if err := runAdd(umbrella, cmd, []string{subPath + "/handler.go"}); err != nil {
		t.Fatalf("runAdd %s/handler.go: %v", subPath, err)
	}

	staged := stagedFilesInDir(t, subDir)
	if !strings.Contains(staged, "handler.go") {
		t.Errorf("expected handler.go to be staged\ngit diff --cached:\n%s", staged)
	}
	if strings.Contains(staged, "other.go") {
		t.Errorf("expected other.go to NOT be staged\ngit diff --cached:\n%s", staged)
	}
}
