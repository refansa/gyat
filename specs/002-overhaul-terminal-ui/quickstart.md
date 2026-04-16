# quickstart.md

This feature provides an interactive terminal UI for repository-oriented commands (`list`, `status`). The interactive UI is the default when the command is run in an interactive terminal, and `--no-ui` forces the existing plain-text output.

Run locally (interactive):

1. Build: `go build ./...`
2. Run `gyat list` — a grouped left sidebar opens by default and the current repository is highlighted.
3. Run `gyat status` — the same sidebar opens with per-repository tabs on the right (`Overview`, `Staged`, `Unstaged`, `Untracked`).

Non-interactive / legacy plain-text output (for scripts, CI, or opt-out):

1. `gyat list --no-ui` — outputs the existing plain-text table suitable for scripts. Example:

```
NAME       BRANCH        STATUS                PATH
frontend   master        clean                 /home/alice/workspace/frontend
backend    feature/x     modified (2 files)    /home/alice/workspace/backend
tooling    develop       diverged (behind 1)   /home/alice/workspace/tooling
```

2. `gyat status <repo> --no-ui` — retains the current plain-text status output (no UI formatting change). The output remains the same as the existing `gyat status` / `git status` style the project currently uses.

Notes:

- Keyboard: Up/Down/PageUp/PageDown/Home/End for navigation; Tab/Shift-Tab and Left/Right to switch tabs; `?` or `h` for help.
- Grouping: repositories are grouped by the first manifest `groups` entry; ungrouped repositories appear under `Other`.
- Resizing: The UI will reflow on terminal resize; extremely narrow terminals will show a compact layout.
- Testing & automation: Because `--json` is not provided in this iteration, use the plain-text `--no-ui` outputs (stable columns) for simple scripting, or prefer the programmatic test harness and unit tests that exercise the Bubble Tea model for robust automation.
