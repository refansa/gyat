# research.md

## Decisions

- Decision: Use Go + Bubble Tea (github.com/charmbracelet/bubbletea) for the interactive terminal UI.
  Rationale: The repository already depends on Bubble Tea, Bubbles, and Lipgloss (see go.mod). These libraries provide an established pattern for list/table views, tabs, and responsive terminal rendering.

- Decision: Preserve git as the source of truth and keep VCS operations delegated to the existing codepaths.
  Rationale: Constitution requires a thin git surface. UI should only display results from existing git invocations and not reimplement VCS behavior.

- Decision: Deterministic ordering: Present repositories in the order defined by the .gyat manifest when present, otherwise fall back to lexicographic path order.
  Rationale: Determinism is required by the Constitution. Using the manifest honors user intent for workspace ordering.

-- Decision (revised): Do not add a `--json` flag at this time. Instead, provide a TTY-driven test harness and unit tests for the Bubble Tea model to validate behavior programmatically.
  Rationale: The team chose to defer machine-readable output to a later iteration; mitigations include test harnesses and stable model APIs for automation.

- Decision: Performance & responsiveness goals: aim for interactive latency <50ms for navigation on workspaces with up to 200 repositories; rendering should avoid full re-renders where possible and use incremental updates via Bubble Tea's Model/Update architecture.
  Rationale: Reasonable responsiveness target for terminal interactions; avoid perceptible lag during navigation.

## Alternatives Considered

- Alternative: Implement a custom TUI engine. Rejected because Bubble Tea already provides battle-tested patterns and is an existing dependency.

- Alternative: Keep pager-style output but add keyboard shortcuts. Rejected because user explicitly requested an interactive list and tabs.

## Open Questions / NEEDS CLARIFICATION (resolved here)

- Terminal resizing behavior: Use Bubble Tea's built-in window size messages to reflow the list and collapse columns when width is constrained. For very narrow widths, show a compact single-column list and allow horizontal scrolling for long labels.

- Keyboard conventions: Use standard navigation keys: up/down, page up/down, home/end, and number keys to jump to entries; use tab/shift-tab and left/right to switch status tabs. Provide keymap help via `?` or `h`.

- Testing approach: Add unit tests for ordering and selection behavior and integration tests that run commands in non-UI (`--no-ui`) mode to assert plain-text outputs. Add a TTY-driven interactive smoke test script for manual QA and model-level unit tests for interactive behavior.

## Implementation Notes & Best Practices

- Use Bubbles' list or table components for repository listing (grouped by manifest `group`) and a custom tab view for status panes; implement a left-sidebar + right content layout matching the provided reference image.
- Avoid blocking I/O in the render/update loop; fetch git data asynchronously and send messages to the Bubble Tea model when ready.
-- Provide a `--no-ui` flag or a programmatic test harness for automation and CI usage (no `--json` at this time).

---

This research resolves the NEEDS CLARIFICATION entries in plan.md: Technology choice, ordering determinism, machine-readable output, performance targets, and terminal resizing/keyboard conventions.
