package cmd

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestResolvePagerCommand_UsesPAGER(t *testing.T) {
	t.Parallel()

	pager, ok := resolvePagerCommand(func(string) (string, bool) {
		return "delta --paging=never", true
	}, "linux")
	if !ok {
		t.Fatal("expected pager command to be resolved")
	}
	if pager.name != "delta" {
		t.Fatalf("expected pager name delta, got %q", pager.name)
	}
	if len(pager.args) != 1 || pager.args[0] != "--paging=never" {
		t.Fatalf("expected pager args [--paging=never], got %v", pager.args)
	}
}

func TestResolvePagerCommand_EmptyPAGERDisablesPaging(t *testing.T) {
	t.Parallel()

	if _, ok := resolvePagerCommand(func(string) (string, bool) {
		return "   ", true
	}, "linux"); ok {
		t.Fatal("expected empty PAGER to disable paging")
	}
}

func TestResolvePagerCommand_DefaultsByPlatform(t *testing.T) {
	t.Parallel()

	t.Run("windows", func(t *testing.T) {
		t.Parallel()

		pager, ok := resolvePagerCommand(func(string) (string, bool) {
			return "", false
		}, "windows")
		if !ok {
			t.Fatal("expected default pager for windows")
		}
		if pager.name != "more" || len(pager.args) != 0 {
			t.Fatalf("expected windows pager 'more', got %#v", pager)
		}
	})

	t.Run("posix", func(t *testing.T) {
		t.Parallel()

		pager, ok := resolvePagerCommand(func(string) (string, bool) {
			return "", false
		}, "linux")
		if !ok {
			t.Fatal("expected default pager for posix")
		}
		if pager.name != "less" {
			t.Fatalf("expected posix pager less, got %q", pager.name)
		}
		if len(pager.args) != 1 || pager.args[0] != "-FRX" {
			t.Fatalf("expected posix pager args [-FRX], got %v", pager.args)
		}
	})
}

func TestWriteMaybePagedOutput_BypassesPagerForNonTerminal(t *testing.T) {
	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return false }
	pagerLookupEnv = func(string) (string, bool) {
		t.Fatal("pager env lookup should not run for non-terminal output")
		return "", false
	}
	pagerRunner = func(io.Writer, io.Writer, string, pagerCommand) error {
		t.Fatal("pager runner should not run for non-terminal output")
		return nil
	}

	var stdout bytes.Buffer
	if err := writeMaybePagedOutput(&stdout, io.Discard, "status output\n", false); err != nil {
		t.Fatalf("writeMaybePagedOutput: %v", err)
	}
	if stdout.String() != "status output\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestWriteMaybePagedOutput_UsesPagerForInteractiveOutput(t *testing.T) {
	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		return "less -FRX", true
	}

	called := false
	var got pagerCommand
	var gotContent string
	pagerRunner = func(stdout, stderr io.Writer, content string, pager pagerCommand) error {
		called = true
		got = pager
		gotContent = content
		return nil
	}

	var stdout bytes.Buffer
	if err := writeMaybePagedOutput(&stdout, io.Discard, "status output\n", false); err != nil {
		t.Fatalf("writeMaybePagedOutput: %v", err)
	}
	if !called {
		t.Fatal("expected pager runner to be called")
	}
	if got.name != "less" || len(got.args) != 1 || got.args[0] != "-FRX" {
		t.Fatalf("unexpected pager command: %#v", got)
	}
	if gotContent != "status output\n" {
		t.Fatalf("unexpected pager content: %q", gotContent)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected pager path to avoid direct stdout writes, got %q", stdout.String())
	}
}

func TestWriteMaybePagedOutput_FallsBackWhenPagerFails(t *testing.T) {
	oldDetector := pagerTerminalDetector
	oldLookup := pagerLookupEnv
	oldRunner := pagerRunner
	t.Cleanup(func() {
		pagerTerminalDetector = oldDetector
		pagerLookupEnv = oldLookup
		pagerRunner = oldRunner
	})

	pagerTerminalDetector = func(io.Writer) bool { return true }
	pagerLookupEnv = func(string) (string, bool) {
		return "less -FRX", true
	}
	pagerRunner = func(io.Writer, io.Writer, string, pagerCommand) error {
		return errors.New("pager failed")
	}

	var stdout bytes.Buffer
	if err := writeMaybePagedOutput(&stdout, io.Discard, "status output\n", false); err != nil {
		t.Fatalf("writeMaybePagedOutput: %v", err)
	}
	if stdout.String() != "status output\n" {
		t.Fatalf("expected fallback stdout write, got %q", stdout.String())
	}
}
