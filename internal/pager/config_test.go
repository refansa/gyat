package pager

import (
	"os"
	"testing"
)

func TestGYATNoPager(t *testing.T) {
	t.Parallel()

	// Ensure empty/unset yields false
	_ = os.Unsetenv("GYAT_NO_PAGER")
	if GYATNoPager() {
		t.Fatalf("expected GYATNoPager false when unset")
	}

	// Truthy values should return true
	_ = os.Setenv("GYAT_NO_PAGER", "1")
	if !GYATNoPager() {
		t.Fatalf("expected GYATNoPager true for '1'")
	}
	_ = os.Setenv("GYAT_NO_PAGER", "true")
	if !GYATNoPager() {
		t.Fatalf("expected GYATNoPager true for 'true'")
	}

	// Non-truthy defaults to false
	_ = os.Setenv("GYAT_NO_PAGER", "nope")
	if GYATNoPager() {
		t.Fatalf("expected GYATNoPager false for 'nope'")
	}
}
