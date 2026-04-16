# gyat Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-15

## Active Technologies
- Go 1.26 + github.com/charmbracelet/bubbletea, github.com/charmbracelet/bubbles, github.com/charmbracelet/lipgloss, golang.org/x/term (001-replace-pager-with-tui)
- N/A (in-memory / file-backed buffer as fallback) (001-replace-pager-with-tui)
- Go 1.26 + github.com/charmbracelet/bubbletea, github.com/charmbracelet/bubbles, github.com/charmbracelet/lipgloss, github.com/spf13/cobra, golang.org/x/term (002-overhaul-terminal-ui)
- N/A (UI-only; no new persistent storage required) (002-overhaul-terminal-ui)

- [e.g., Python 3.11, Swift 5.9, Rust 1.75 or NEEDS CLARIFICATION] + [e.g., FastAPI, UIKit, LLVM or NEEDS CLARIFICATION] (001-windows-pager)

## Project Structure

```text
backend/
frontend/
tests/
```

## Commands

cd src; pytest; ruff check .

## Code Style

[e.g., Python 3.11, Swift 5.9, Rust 1.75 or NEEDS CLARIFICATION]: Follow standard conventions

## Recent Changes
- 002-overhaul-terminal-ui: Added Go 1.26 + github.com/charmbracelet/bubbletea, github.com/charmbracelet/bubbles, github.com/charmbracelet/lipgloss, github.com/spf13/cobra, golang.org/x/term
- 001-replace-pager-with-tui: Added Go 1.26 + github.com/charmbracelet/bubbletea, github.com/charmbracelet/bubbles, github.com/charmbracelet/lipgloss, golang.org/x/term

- 001-windows-pager: Added [e.g., Python 3.11, Swift 5.9, Rust 1.75 or NEEDS CLARIFICATION] + [e.g., FastAPI, UIKit, LLVM or NEEDS CLARIFICATION]

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
