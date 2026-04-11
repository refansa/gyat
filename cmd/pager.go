package cmd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const noPagerFlagName = "no-pager"

type pagerCommand struct {
	name string
	args []string
}

var pagerLookupEnv = os.LookupEnv
var pagerLookPath = exec.LookPath
var pagerTerminalDetector = defaultPagerTerminalDetector
var pagerRunner = runPagerCommand

func bindNoPagerFlag(command *cobra.Command) {
	command.Flags().Bool(noPagerFlagName, false, "Disable paging even when stdout is an interactive terminal")
}

func noPagerEnabled(command *cobra.Command) bool {
	if command == nil {
		return false
	}
	if command.Flags().Lookup(noPagerFlagName) == nil {
		return false
	}
	enabled, err := command.Flags().GetBool(noPagerFlagName)
	return err == nil && enabled
}

func writeMaybePagedOutput(stdout, stderr io.Writer, content string, disabled bool) error {
	if content == "" {
		return nil
	}

	pager, ok := activePagerCommand(stdout, disabled)
	if !ok {
		_, err := io.WriteString(stdout, content)
		return err
	}

	if err := pagerRunner(stdout, stderr, content, pager); err != nil {
		_, writeErr := io.WriteString(stdout, content)
		return writeErr
	}

	return nil
}

func activePagerCommand(stdout io.Writer, disabled bool) (pagerCommand, bool) {
	if disabled || !pagerTerminalDetector(stdout) {
		return pagerCommand{}, false
	}

	return resolvePagerCommand(pagerLookupEnv, pagerLookPath, runtime.GOOS)
}

func resolvePagerCommand(lookupEnv func(string) (string, bool), lookPath func(string) (string, error), goos string) (pagerCommand, bool) {
	if pagerValue, ok := lookupEnv("PAGER"); ok {
		return parsePagerCommand(pagerValue)
	}

	return defaultPagerCommand(lookPath, goos), true
}

func parsePagerCommand(value string) (pagerCommand, bool) {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return pagerCommand{}, false
	}

	return pagerCommand{name: fields[0], args: fields[1:]}, true
}


func defaultPagerCommand(lookPath func(string) (string, error), goos string) pagerCommand {
	if goos == "windows" {
		if _, err := lookPath("less"); err == nil {
			return pagerCommand{name: "less", args: []string{"-FRX"}}
		}
		return pagerCommand{name: "more"}
	}

	return pagerCommand{name: "less", args: []string{"-FRX"}}
}

func pagerUsesASCIIStyle(pager pagerCommand) bool {
	name := strings.ToLower(filepath.Base(strings.TrimSpace(pager.name)))
	return name == "more" || name == "more.com"
}

func defaultPagerTerminalDetector(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func runPagerCommand(stdout, stderr io.Writer, content string, pager pagerCommand) error {
	if _, err := exec.LookPath(pager.name); err != nil {
		return err
	}

	command := exec.Command(pager.name, pager.args...)
	command.Stdin = strings.NewReader(content)
	command.Stdout = stdout
	command.Stderr = stderr

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) || errors.Is(err, syscall.EPIPE) {
			return nil
		}
		return err
	}

	return nil
}
