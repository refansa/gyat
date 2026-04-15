package cmd

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestWriteMaybePagedOutput_UsesInternalTUIForInteractiveTTY(t *testing.T) {
	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	oldTUIRunner := pagerTUIRunner
	oldStdin := pagerStdin
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
		pagerTUIRunner = oldTUIRunner
		pagerStdin = oldStdin
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		t.Fatal("external pager should not be resolved when the internal TUI succeeds")
		return "", false
	}
	pagerRunner = func(io.Writer, io.Writer, string, pagerCommand) error {
		t.Fatal("external pager should not run when the internal TUI succeeds")
		return nil
	}

	stdinFile, err := os.CreateTemp(t.TempDir(), "pager-stdin")
	if err != nil {
		t.Fatalf("CreateTemp stdin: %v", err)
	}
	defer stdinFile.Close()
	pagerStdin = stdinFile

	stdoutFile, err := os.CreateTemp(t.TempDir(), "pager-stdout")
	if err != nil {
		t.Fatalf("CreateTemp stdout: %v", err)
	}
	defer stdoutFile.Close()

	called := false
	pagerTUIRunner = func(content []byte, in *os.File, out *os.File) error {
		called = true
		if string(content) != "status output\n" {
			t.Fatalf("unexpected content: %q", string(content))
		}
		if in != stdinFile || out != stdoutFile {
			t.Fatalf("unexpected TUI files: in=%v out=%v", in.Name(), out.Name())
		}
		return nil
	}

	if err := writeMaybePagedOutput(stdoutFile, io.Discard, "status output\n", false); err != nil {
		t.Fatalf("writeMaybePagedOutput: %v", err)
	}
	if !called {
		t.Fatal("expected internal TUI runner to be called")
	}
}

func TestWriteMaybePagedOutput_FallsBackWhenInternalTUIFails(t *testing.T) {
	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	oldTUIRunner := pagerTUIRunner
	oldStdin := pagerStdin
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
		pagerTUIRunner = oldTUIRunner
		pagerStdin = oldStdin
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		return "less -FRX", true
	}

	stdinFile, err := os.CreateTemp(t.TempDir(), "pager-stdin")
	if err != nil {
		t.Fatalf("CreateTemp stdin: %v", err)
	}
	defer stdinFile.Close()
	pagerStdin = stdinFile

	stdoutFile, err := os.CreateTemp(t.TempDir(), "pager-stdout")
	if err != nil {
		t.Fatalf("CreateTemp stdout: %v", err)
	}
	defer stdoutFile.Close()

	pagerTUIRunner = func([]byte, *os.File, *os.File) error {
		return errors.New("tui failed")
	}

	called := false
	pagerRunner = func(stdout, stderr io.Writer, content string, pager pagerCommand) error {
		called = true
		if pager.name != "less" {
			t.Fatalf("unexpected pager: %#v", pager)
		}
		if content != "status output\n" {
			t.Fatalf("unexpected content: %q", content)
		}
		return nil
	}

	if err := writeMaybePagedOutput(stdoutFile, io.Discard, "status output\n", false); err != nil {
		t.Fatalf("writeMaybePagedOutput: %v", err)
	}
	if !called {
		t.Fatal("expected external pager fallback after TUI failure")
	}
}
