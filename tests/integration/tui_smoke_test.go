package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTUISmokeHarness(t *testing.T) {
	if os.Getenv("GYAT_RUN_TUI_SMOKE") == "" {
		t.Skip("manual smoke test: set GYAT_RUN_TUI_SMOKE=1 and run on an interactive terminal")
	}

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./...: %v\n%s", err, output)
	}
	// The interactive browser still requires a real terminal; the gated build step
	// keeps this harness useful in CI and manual verification.
}
