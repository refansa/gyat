package cmd

import (
	"bytes"
	"io"
	"testing"
)

// Test that writing to a pipe (simulating piping stdout to another process)
// preserves the raw bytes produced by the report — i.e., the consumer gets the
// same bytes that would be written when paging is bypassed.
func TestWriteMaybePagedOutput_PipePreservesBytes(t *testing.T) {
	t.Parallel()

	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
	})

	// Simulate a pipe by reporting the writer is not a terminal.
	pagerTerminalDetector = func(io.Writer) bool { return false }
	pagerLookupEnv = func(string) (string, bool) {
		t.Fatal("pager env lookup should not run for pipe output")
		return "", false
	}
	pagerRunner = func(io.Writer, io.Writer, string, pagerCommand) error {
		t.Fatal("pager runner should not run for pipe output")
		return nil
	}

	r, w := io.Pipe()
	defer r.Close()

	content := []byte("line1\nline2\n")
	// Start writer goroutine
	writeErrCh := make(chan error, 1)
	go func() {
		writeErrCh <- writeMaybePagedOutput(w, io.Discard, string(content), false)
		_ = w.Close()
	}()

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read from pipe: %v", err)
	}

	if err := <-writeErrCh; err != nil {
		t.Fatalf("writeMaybePagedOutput returned error: %v", err)
	}

	if !bytes.Equal(got, content) {
		t.Fatalf("piped bytes differ: got=%q want=%q", string(got), string(content))
	}
}
