Decision: Use Bubble Tea (charmbracelet) for the interactive TUI viewer

Rationale:
- Bubble Tea is a mature, well-documented Go TUI framework with an active ecosystem (Bubbles for components, Lipgloss for styling). It provides a clear MVU model that suits single-screen interactive viewers and handles cross-platform terminal quirks.
- It integrates well with existing Go code and avoids adding a separate binary or runtime. The repo already targets Go 1.26 which is compatible with Bubble Tea.
- The framework is widely used and has patterns for efficient view updates and virtualization which help with large-output performance.

Alternatives considered:
- Implement custom terminal UI using golang.org/x/term and hand-rolled rendering: rejected due to higher maintenance cost and subtle cross-platform issues.
- Use an external pager binary (less, more) exclusively: rejected because it limits keyboard-driven discoverability and consistent in-view search behavior across platforms.

Platform notes:
- Windows: Bubble Tea works on Windows terminals; keep the existing Windows-specific RunInteractiveSession as a minimal fallback if necessary. Where possible, implement platform-agnostic code and test on Windows terminals.

Performance considerations and best practices:
- Avoid materializing extremely large content in memory if possible; instead keep lines as byte slices and render only the visible window using a viewport or lazy rendering.
- Use virtualization: only compute visible lines and search matches on demand. For initial MVP, keep a simple in-memory slice of strings and optimize if profiling shows issues.
- For search, avoid full regex by default; implement substring search first. Consider an index for very large inputs as a later optimization.

Testing & automation:
- Non-interactive behavior must remain unchanged. Add integration tests that simulate stdout piped scenarios and confirm the output is identical.
- Unit tests: test model Update functions, viewport rendering, search indexing, and key handling.

Security/constraints:
- No network calls. No external resources. Keep the dependency set minimal (bubbletea, bubbles, lipgloss) and review transitive dependencies.

Decision Summary:
- Selected: github.com/charmbracelet/bubbletea + Bubbles + Lipgloss
- Why: Good ergonomics for Go, cross-platform support, community adoption, aligns with repo conventions and language version.
