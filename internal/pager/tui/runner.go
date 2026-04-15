package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// RunTUI starts the Bubble Tea pager session.
func RunTUI(content []byte, in *os.File, out *os.File) error {
	if in == nil || out == nil {
		return fmt.Errorf("stdin/stdout must be provided")
	}
	if !term.IsTerminal(int(in.Fd())) || !term.IsTerminal(int(out.Fd())) {
		return fmt.Errorf("stdin/stdout must be terminals")
	}

	width, height := 80, 24
	if w, h, err := term.GetSize(int(out.Fd())); err == nil {
		width = w
		height = h
	}

	program := tea.NewProgram(
		NewModel(content, width, height),
		tea.WithInput(in),
		tea.WithOutput(out),
		tea.WithAltScreen(),
	)

	_, err := program.Run()
	return err
}
