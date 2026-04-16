package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
	"golang.org/x/term"
)

// Run starts the repository browser in a Bubble Tea alt-screen session.
func Run(title string, entries []uiModel.RepositoryEntry, in *os.File, out *os.File) error {
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
		NewModel(title, entries, width, height),
		tea.WithInput(in),
		tea.WithOutput(out),
		tea.WithAltScreen(),
	)

	_, err := program.Run()
	return err
}
