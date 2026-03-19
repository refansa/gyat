# Copilot Instructions

## Project Overview

`gyat` (Git Your Ass Together) is a CLI tool written in Go that manages git submodules across
multiple related repositories. It is a thin, opinionated wrapper around `git submodule` commands,
designed to feel familiar to anyone who already uses git.

The guiding principle: **gyat should feel like a natural extension of git, not a separate tool.**

---

## Build & Run

```sh
go build ./...
go run . <command>
```

Build the binary:
```sh
go build -o gyat .
```

Install the binary to `$GOPATH/bin`:
```sh
go install .
```

---

## Tests

Tests exist across two packages. Run them with:

```sh
go test ./...                              # full suite
go test ./cmd/... -v                       # cmd package, verbose
go test ./cmd/... -run TestRunAdd          # single test by name prefix
go test ./internal/git/... -v              # git runner tests only
```

Test layout:
- `internal/git/git_test.go` — unit tests for `Run` and error wrapping behaviour
- `cmd/testhelper_test.go`   — shared helpers: repo setup, stdout/stderr capture, assertions
- `cmd/<command>_test.go`    — per-command tests (unit + integration)
- `cmd/integration_test.go`  — full end-to-end workflow tests (init → add → list → remove)

Integration tests spin up real git repositories inside `t.TempDir()` directories and call the
`run*` functions directly. They do not touch the network or any repository outside the temp dir.
All temp directories are cleaned up automatically by the test framework.

When adding tests for a new command, follow the pattern in the existing `cmd/<command>_test.go`
files: use `newTestSetup(t, name)` or `newUmbrellaRepo(t)` from the shared helper, `t.Chdir()`
into the test repo, and call `run<Command>()` directly rather than going through cobra.

---

## Architecture

```
main.go                  → calls cmd.Execute(), prints errors to stderr, exits with code 1
cmd/root.go              → defines rootCmd, registers all subcommands in init()
cmd/<command>.go         → one file per subcommand (add, commit, init, list, remove, sync, update)
internal/git/git.go      → the only git abstraction: Run() and RunInteractive()
```

### The two execution modes in `internal/git`

- `git.Run(args...)` — captures stdout, returns `(string, error)`. Stderr is folded into the
  error on failure. Use this when the output needs to be parsed (e.g. reading `.gitmodules`
  config, parsing `submodule status`).

- `git.RunInteractive(args...)` — passes stdin/stdout/stderr straight through to the terminal.
  Use this when the command produces live output, progress bars, or credential prompts
  (e.g. `add`, `update`, `sync`).

### Per-command file structure

Each command file follows the same pattern:

1. Declare `var <name>Cmd *cobra.Command` with `Use`, `Short`, `Long`, `Args`, and `RunE`.
2. Implement `func run<Name>(cmd *cobra.Command, args []string) error` — builds git args and
   delegates to `git.Run` or `git.RunInteractive`.
3. Register flags bound to package-level vars in `func init()`.
4. Register the command on `rootCmd` in `cmd/root.go`'s `init()`.

### Shared helpers across commands

Some functions are shared between command files within the `cmd` package:

- `allSubmodulePaths(dir string) ([]string, error)` — reads every submodule path from
  `.gitmodules`. Used by `add` and `commit`.
- `hasStagedChanges(statusOut string) bool` — checks `git status --porcelain` output for
  staged (index) changes. Defined in `commit.go`.
- `hasWorkingTreeChanges(statusOut string) bool` — checks porcelain output for unstaged
  working-tree changes. Defined in `add.go`.

---

## Output & UX Conventions

gyat output must feel consistent with git. Follow these rules for every command:

### stdout vs stderr

| What                              | Where    |
|-----------------------------------|----------|
| Data output (e.g. `list` table)   | stdout   |
| "No data" responses               | stdout   |
| Progress messages                 | stderr   |
| Completion messages               | stderr   |
| Warnings                          | stderr   |
| Hints / suggestions               | stderr   |
| Errors (via `RunE` return)        | stderr (cobra handles this) |

The rule: if the user might pipe the output, it goes to stdout. Everything else goes to stderr.

```go
// Correct
fmt.Fprintln(os.Stderr, "removing submodule 'auth'...")
fmt.Println(tableRow)

// Wrong
fmt.Println("removing submodule 'auth'...")     // progress on stdout
fmt.Fprintln(os.Stderr, tableRow)               // data on stderr
```

### Message style

Follow git's style exactly:

| Type       | Format                                                              | Example                                                         |
|------------|---------------------------------------------------------------------|-----------------------------------------------------------------|
| Progress   | lowercase, ends with `...`                                          | `syncing submodule URLs...`                                     |
| Completion | lowercase, no trailing punctuation                                  | `removed submodule 'auth'`                                      |
| Warning    | `warning:` prefix, lowercase                                        | `warning: 'path' is an absolute path and will only work on this machine` |
| Hint       | `hint:` prefix, lowercase, no trailing punctuation                  | `hint: use 'gyat add <repo>' to start adding repositories`      |
| Fatal      | return an `error` — cobra prints it prefixed with `Error:` already  | `return fmt.Errorf("failed to deinit submodule: %w", err)`      |

Never use:
- Sentence-case or title-case for progress/completion/hint messages
- Trailing periods on progress or hint messages
- Emoji or Unicode symbols (em dashes `—`, arrows `→`, and box-drawing chars `─` are fine)

---

## Key Conventions

- **No git library dependency.** All git operations go through `internal/git` by shelling out to
  the `git` binary. Do not introduce a Go git library (e.g. `go-git`).

- **Errors bubble via `RunE`.** Commands use `RunE` (not `Run`) so errors propagate naturally to
  `main.go`, which prints to stderr and exits with code 1.

- **Local path transport fix.** Git 2.38.1+ blocks local file transport by default
  (CVE-2022-39253). When `isLocalPath(source)` is true, prepend
  `-c protocol.file.allow=always` to the git args before `submodule add`.

- **`remove` does the full three-step cleanup** manually:
  1. `git submodule deinit -f <path>`
  2. `os.RemoveAll(.git/modules/<path>)`
  3. `git rm -f <path>`
  This is intentional — git has no single command for clean submodule removal.

- **`list` parses `.gitmodules` directly** via `git config -f .gitmodules` rather than relying
  on porcelain output, so it can surface URL and branch metadata alongside status.

- **`commit` cascades from submodules to umbrella.** It first commits each submodule that has
  staged changes, then stages the updated submodule refs (`git add <path>`) in the umbrella
  repository, and finally commits the umbrella — all with the same message. When path
  arguments are provided, only the named submodules are committed. The `runCommit` function
  accepts `message` and `noVerify` as explicit parameters (not from global flag state) so
  tests can run in parallel without races, following the same pattern as `runTrack`.

- **New subcommands** belong in `cmd/<name>.go` and must be registered in `cmd/root.go`'s
  `init()`.