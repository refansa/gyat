# Implementation Plan: Overhaul Terminal UI

**Branch**: `002-overhaul-terminal-ui` | **Date**: 2026-04-15 | **Spec**: `specs/002-overhaul-terminal-ui/spec.md`
**Input**: Feature specification from `/specs/002-overhaul-terminal-ui/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Replace pager-style repository output for `list` and `status` with an interactive Bubble Tea UI. The implementation uses a grouped repository sidebar, a right-hand tabbed detail pane, shared UI components reused across commands, and `--no-ui` plus automatic non-TTY fallback for plain-text output.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.26
**Primary Dependencies**: github.com/charmbracelet/bubbletea, github.com/charmbracelet/bubbles, github.com/charmbracelet/lipgloss, github.com/spf13/cobra, golang.org/x/term
**Storage**: N/A (UI-only; no new persistent storage required)
**Testing**: Go testing (`go test`) with model/component tests in `internal/ui/tui`, command routing tests in `cmd`, and manual smoke/integration stubs in `tests/integration`
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: CLI application (terminal UI)
**Performance Goals**: Interactive navigation target under 50ms for workspaces up to 200 repositories; benchmark coverage added for model navigation
**Constraints**: Must remain a thin wrapper around git, preserve non-destructive defaults, be keyboard-first, and automatically fall back to plain-text output when stdout is not a TTY or when `--no-ui` is set
**Scale/Scope**: Designed for workspaces with tens to hundreds of repositories; acceptance criteria reference tests with >=20 repositories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

[Conformance summary based on .specify/memory/constitution.md]

1. Workspace-Centric Model: This feature modifies only the presentation layer (terminal UI) and does not alter the `.gyat` manifest or tracked repositories. No migration of manifest required. ✓

2. Thin Git Surface: The redesign must remain a thin wrapper around the git binary; UI changes must delegate VCS semantics to git and not reimplement git operations. Implementation will keep existing git command invocations. ✓

3. Deterministic Execution: CLI behavior that iterates repositories remains deterministic. The interactive view uses manifest order when known and lexicographic path order as a fallback, with tests covering both cases. ✓

4. Safe, Non-Destructive Defaults: The feature is UI-only and should not change defaults for destructive operations. Confirmed. ✓

5. CLI-First, Observable, Testable: The Constitution recommends machine-readable outputs for automation. Decision: the team will not add a `--json` flag at this time (documented deviation). Mitigation: provide a TTY-driven test harness and unit tests for model logic; ensure output stream conventions (data→stdout, progress/warnings→stderr) remain consistent. This deviation is recorded in Complexity Tracking below and must be re-evaluated after Phase 1 design.

Gate result (pre-research): PASS with one documented deviation (no --json) and one resolved clarification (ordering determinism). The JSON/machine-readable requirement is a conscious, documented deviation and appears in Complexity Tracking. This must be reviewed after Phase 1 design; unresolved constitutional concerns must be escalated.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: Single Go CLI project. Interactive UI code lives in `internal/ui/model`, `internal/ui/data`, `internal/ui/components`, and `internal/ui/tui`; command integration remains in `cmd/`; manual smoke/integration coverage remains in `tests/integration/`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Deviation: No machine-readable `--json` output | The team chose to prioritize interactive TUI parity with the reference UI and defer `--json` until after initial rollout. | Machine-readable output aids automation; however, adding and testing an interactive UI first reduces early complexity and implementation scope. Mitigation: provide a TTY-driven test harness and unit tests for model logic. |
