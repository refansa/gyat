# Implementation Plan: Replace pager with a Bubble Tea-based TUI viewer

**Branch**: `001-replace-pager-with-tui` | **Date**: 2026-04-15 | **Spec**: /specs/001-replace-pager-with-tui/spec.md
**Input**: Feature specification from `/specs/001-replace-pager-with-tui/spec.md`

## Summary

Replace the existing lightweight pager with a keyboard-driven, Bubble Tea (charmbracelet) based TUI viewer that provides predictable line/page navigation, in-view search with match navigation, and a visible position indicator. The change must remain backwards-compatible for non-interactive workflows (no TUI invoked when stdout is not a TTY or when --no-pager / GYAT_NO_PAGER is set) and preserve exit codes and piped output.

## Technical Context

**Language/Version**: Go 1.26  
**Primary Dependencies**: github.com/charmbracelet/bubbletea, github.com/charmbracelet/bubbles, github.com/charmbracelet/lipgloss, golang.org/x/term  
**Storage**: N/A (in-memory / file-backed buffer as fallback)  
**Testing**: go test (unit and integration); existing tests under tests/ will be extended to cover the TUI behavior  
**Target Platform**: CLI (Linux, macOS, Windows — cross-platform terminal)  
**Project Type**: CLI (internal library: internal/pager)  
**Performance Goals**: Viewer opens and becomes interactive for outputs up to 50,000 lines within ~2 seconds on a typical developer laptop; memory usage should remain reasonable (target <200MB) and the viewer must degrade gracefully for extremely large inputs.  
**Constraints**: Must preserve automation: non-interactive contexts (stdout not a TTY or piped) must emit raw output unchanged. Must avoid introducing persistent credential handling or network calls.  
**Scale/Scope**: Single binary; change scoped to internal/pager and cmd integration with minimal surface area updates to avoid broad regressions.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Gates (from .specify/memory/constitution.md) and evaluation:

- Language & Tooling: gyat is implemented in Go and targets Go 1.26 — Decision: Use Go 1.26 (PASS)
- CLI-First & Output Conventions: Viewer must not break non-interactive behavior and must preserve stdout/stderr conventions — Decision: Detect TTY and respect --no-pager / GYAT_NO_PAGER; write raw output when not interactive (PASS)
- Safe, Non-Destructive Defaults: Feature will default to non-destructive, opt-out via flags/env — Decision: honor environment and flags; no destructive changes (PASS)
- Deterministic Execution: Viewer will not change command-side ordering; any parallelization is out-of-scope (PASS)

No constitution gates are violated by this proposal. If later a gate is at risk (for example if a new dependency introduced security concerns), the design will be revisited and documented in the Complexity Tracking table.

## Project Structure

### Documentation (this feature)

```text
specs/001-replace-pager-with-tui/
├── plan.md              # This file
├── research.md          # Phase 0 output (this plan run)
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (pager contract)
└── tasks.md             # Phase 2 output (generated later)
```

### Source Code (repository root)

```text
internal/pager/         # existing pager package
├── config.go
├── doc.go
├── output.go
├── pager.go             # retained API: Render, HandleKey, Search, Seek
├── session.go
├── tty.go
├── windows_windows.go   # existing interactive loop for Windows
└── tui/                 # new package/files implementing Bubble Tea viewer
    ├── model.go         # bubbletea Model and Update/View implementations
    ├── viewport.go      # helpers for efficient viewport rendering
    └── style.go         # optional lipgloss styles

cmd/                    # existing CLI integration (cmd/pager.go will be adapted)
```

**Structure Decision**: Keep the public/internal API used by cmd/pager.go stable (Pager type + Render + RunInteractiveSession) and implement a new tui/ subpackage under internal/pager to contain Bubble Tea specific code. This minimizes caller changes and keeps the refactor local to internal/pager.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations detected; no complexity justification required.
