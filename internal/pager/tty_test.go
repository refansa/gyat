package pager

import (
	"os"
	"testing"
)

func TestIsTerminal_NonFile(t *testing.T) {
	t.Parallel()

	// A non-*os.File writer should return false
	var _ = func() {}
	if IsTerminal(nil) {
		t.Fatalf("expected nil writer to not be a terminal")
	}
}

func TestIsTerminal_File(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp("", "pager-tty-test")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Temp files are not character devices; expect false deterministically.
	if IsTerminal(f) {
		t.Fatalf("expected temp file to not be a terminal")
	}
}
