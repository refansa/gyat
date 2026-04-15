---

description: "Task list for Replace Pager With TUI Viewer"
---

# Tasks: Replace Pager With TUI Viewer

**Input**: Design documents from `C:\Users\user\workspace\github.com\refansa\gyat\specs\001-replace-pager-with-tui/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Notes

- All file paths below are absolute to the repository root. Replace `C:\Users\user\workspace\github.com\refansa\gyat` with your local path if different.
- Tests are included for the non-interactive preservation story (US2) because preserving automation is a high-priority acceptance requirement.
- Suggested MVP: deliver User Story 1 (navigation + search + exit) first.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic scaffolding required before implementation starts

- [X] T001 Create package skeleton for the new TUI implementation: create files and basic TODOs at C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\model.go, C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\viewport.go, C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\search.go, C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\runner.go, C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\style.go (each file should declare `package tui` and include a short TODO comment and function skeletons where applicable)
- [X] T002 [P] Add Bubble Tea and styling dependencies to the module: run `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles github.com/charmbracelet/lipgloss golang.org/x/term` from C:\Users\user\workspace\github.com\refansa\gyat and run `go mod tidy` (update go.mod)
- [X] T003 Implement API surface for starting the TUI: in C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\runner.go implement `func RunTUI(content []byte, in *os.File, out *os.File) error` that starts the Bubble Tea program and returns when the user quits (initially a stub that starts the program and returns nil on normal exit)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core integration and safety checks that block user story work

- [X] T004 Update CLI integration to prefer the internal TUI when interactive: modify C:\Users\user\workspace\github.com\refansa\gyat\cmd\pager.go so that when stdout is a TTY and paging is enabled it calls internal/pager/tui.RunTUI; on error or non-supporting environment, fall back to the existing external pager behavior or raw output. Preserve existing env flags (GYAT_NO_PAGER) and --no-pager flag behavior.
- [X] T005 [P] Add or extend integration test to verify non-interactive/piped behavior remains unchanged: create/modify C:\Users\user\workspace\github.com\refansa\gyat\tests\integration\pager_pipe_preserve_test.go to run a command that prints sample content, pipe it, and assert the captured output bytes are identical before and after the change (ensures no TUI invoked when stdout is not a TTY)

---

## Phase 3: User Story 1 - Navigate long CLI output (Priority: P1) 🎯 MVP

**Goal**: Provide a keyboard-driven viewer for multi-screen output with predictable line/page navigation, in-view substring search with match navigation, visible position indicator, and immediate exit returning to the shell prompt.

**Independent Test**: Run a command that prints thousands of lines in an interactive terminal; the TUI viewer must open, support line/page movement (j/k and space/b), support search (invoke with `/`, enter substring, navigate matches with `n`/`N`), display current position (line/total or percentage), and exit with `q` returning to the shell prompt with no extra output.

### Implementation Tasks

- [X] T006 [US1] Implement Bubble Tea model with core state and Update/View methods at C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\model.go. Requirements: hold content lines, cursor/top index, viewport height, matches/search query, implement tea.Model interface, and handle key messages for 'q',' ','b','j','k','n','N','/','?' (help) and terminal resize messages.
- [X] T007 [US1] Implement viewport helpers at C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\viewport.go. Requirements: functions to compute visible line slice given top index and height, ensure safe bounds, and expose helper `RenderLines(lines []string, top, height int) string` for view rendering.
- [X] T008 [US1] Implement substring search and match navigation at C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\search.go. Requirements: function `SearchLines(lines []string, query string) []int` returning match line indexes and helpers `NextMatch(matches []int, currentTop int) (index int)` and `PrevMatch(...)` used by the model.
- [X] T009 [US1] Implement the TUI runner to wire the model and start Bubble Tea at C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\runner.go. Requirements: construct model from provided content, start a bubbletea.Program, block until it exits, and return any program error.
- [X] T010 [US1] Wire the TUI into the CLI flow: adjust C:\Users\user\workspace\github.com\refansa\gyat\cmd\pager.go to call `tui.RunTUI` for interactive sessions (preserving existing fallbacks). Ensure tests still pass and non-interactive code paths are unchanged.
- [X] T011 [US1] Add on-screen help overlay in C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\model.go toggled via '?' or 'h' that shows keybindings (line/page, search, next/prev match, help, exit) with concise instructions.
- [X] T012 [US1] Add a visible position indicator in the viewer (render line/total and percentage) in the rendered view; implement formatting in C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\style.go or model.go as appropriate.
- [X] T013 [P] [US1] Update quickstart and developer docs at C:\Users\user\workspace\github.com\refansa\gyat\specs\001-replace-pager-with-tui\quickstart.md to include interactive usage examples and verification steps for the independent test above

**Checkpoint**: After these tasks, User Story 1 should be functionally complete and manually verifiable in an interactive terminal.

---

## Phase 4: User Story 2 - Preserve non-interactive behavior and automation (Priority: P2)

**Goal**: Ensure existing automation and piping behavior remains unchanged (no TUI invoked, output bytes and exit codes preserved).

**Independent Test**: Run `gyat status | cat > out.txt` or equivalent pipeline and ensure `out.txt` content equals the original output and exit codes are preserved.

### Implementation Tasks

- [X] T014 [US2] Add an integration test that runs the CLI in a non-TTY context and asserts output equality and preserved exit status at C:\Users\user\workspace\github.com\refansa\gyat\tests\integration\pager_noninteractive_preserve_test.go
- [X] T015 [P] [US2] Add a small test helper program to produce large output for manual and CI testing at C:\Users\user\workspace\github.com\refansa\gyat\tools\pager-smoke\main.go (prints N lines; accepts count via flag) and document usage in quickstart.md

---

## Phase 5: User Story 3 - Discoverability and quick help (Priority: P3)

**Goal**: Provide an obvious and discoverable help view that lists the basic controls and is readable in the viewer.

**Independent Test**: Open the TUI viewer and press '?' or 'h' to reveal help text which lists navigation, search, match navigation, and exit controls.

### Implementation Tasks

- [X] T016 [US3] Improve help text styling in C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\tui\style.go and ensure help text is accessible/readable at common terminal sizes
- [X] T017 [US3] Add a small integration manual check in quickstart.md showing the help toggle and sample keybindings to verify discoverability (update the existing quickstart file)

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Performance tuning, CI, documentation, and cleanup

- [X] T018 Add/extend a benchmark to measure viewer initialization and rendering for large inputs at C:\Users\user\workspace\github.com\refansa\gyat\internal\pager\benchmark_test.go (target: measure reasonably for 50,000 lines)
- [X] T019 [P] Run formatting and linting: ensure repository formatted (`gofmt`), run `go vet`, and update README or docs as needed (document commands in C:\Users\user\workspace\github.com\refansa\gyat\specs\001-replace-pager-with-tui\quickstart.md)

---

## Dependencies & Execution Order

- Setup (Phase 1) must complete before Foundational (Phase 2).
- Foundational (Phase 2) must complete before any User Story phases.
- User Story 1 (P1) is the MVP and should be delivered first. User Stories 2 and 3 can be implemented after P1; tests for P2 should be added early to prevent regressions.

### Suggested execution ordering (blocking edges)

1. T001 → T002 → T003 (Setup)
2. T004 → T005 (Foundational)
3. T006 → T007 → T008 → T009 → T010 → T011 → T012 → T013 (US1 implementation)
4. T014 → T015 (US2 testing & helpers)
5. T016 → T017 (US3 discoverability)
6. T018 → T019 (Polish)

## Parallel Opportunities (examples)

- T002 (add deps), T013 (docs), and T015 (smoke helper) are parallelizable [P] and can be worked on concurrently by different developers.
- Within US1, implementing separate files for rendering and search (T007 and T008) can be worked on concurrently if the team agrees on shared interfaces; otherwise follow serial ordering.

---

## Implementation Strategy

1. MVP First: focus on Phase 1 + Phase 2 + Phase 3 (User Story 1) and validate the independent test.

---

## Format Validation

- All tasks listed above follow the required checklist format: each starts with `- [ ]`, includes a Task ID (T001..T019), includes `[P]` only where parallelizable, includes story labels `[US1]`, `[US2]`, or `[US3]` for user story tasks, and contains exact absolute file paths.
