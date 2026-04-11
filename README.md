# gyat - Git Your Ass Together

gyat is an umbrella workspace manager for multi-repository projects. It keeps a
normal git repository at the root, tracks child repositories in a `.gyat`
manifest, and lets you run common git workflows across the workspace without
turning those child repositories into submodules.

## Why?

Sometimes what gets called "microservices" is really a distributed monolith:
10+ repositories that are tightly coupled, deployed together, and changed
together.

If that sounds familiar, `gyat` is for you.

Managing that kind of repo sprawl usually means:

- Jumping between repositories just to finish one feature
- Keeping branches and changes in sync manually
- Opening a pile of PRs for one logical change
- Onboarding people into a workspace that only exists in tribal knowledge

`gyat` gives you one workspace root, one manifest, and one CLI for operating
across the whole set.

### I meant, what with the funny name?

That is obviously the most important bit. Initially I considered naming the
project "Yet Another Git Tracker" (`yagt`), but it felt awkward to type and say.
Rearranging the abbreviation into `gyat` was funnier, shorter, and more honest:
these repositories usually need to get their act together.

## Goals

- Aggregate multiple related repositories under one workspace
- Keep each child repository a normal git repository
- Simplify common multi-repo workflows like `track`, `exec`, `add`, `commit`, `pull`, `push`, `update`, `sync`, and `list`
- Stay thin on top of git rather than replacing it
- Make target selection predictable with repo, group, and root selectors

## Installation

Using `go install` (recommended, requires Go 1.26+):

```sh
go install github.com/refansa/gyat/v2@latest
```

The binary will be placed in `$GOPATH/bin`, so make sure that directory is on
your `PATH`.

From source (requires Go 1.26+):

```sh
git clone https://github.com/refansa/gyat
cd gyat
go build -o gyat .
```

Then move the binary somewhere on your `PATH`:

```sh
# Linux / macOS
mv gyat /usr/local/bin/

# Windows (PowerShell, run as Administrator)
Move-Item gyat.exe C:\Windows\System32\gyat.exe
```

Verify the install:

```sh
gyat --help
```

## Workspace model

Every gyat workspace has a `.gyat` manifest at the umbrella root:

```json
{
	"version": 1,
	"ignore": [
		"node_modules",
		".vscode"
	],
	"repos": [
		{
			"name": "auth",
			"path": "services/auth",
			"url": "git@github.com:org/service-auth.git",
			"branch": "main",
			"groups": ["backend"]
		}
	]
}
```

The model is simple:

- The umbrella root stays a normal git repository.
- Each tracked child repository stays a normal git repository cloned inside the workspace.
- `.gyat` is the source of truth for tracked repository paths, remotes, branches, and groups.
- The root `.gitignore` gets a gyat-managed block for tracked repo paths and manifest ignore patterns.

## Common target flags

Most workspace-aware commands share the same target selection flags:

| Flag | Meaning |
|------|---------|
| `--repo <name-or-path>` | Select one or more tracked repos by name or manifest path |
| `--group <group>` | Select repos whose `.gyat` entry contains that group |
| `--no-root` | Exclude the umbrella repository |
| `--root-only` | Target only the umbrella repository |
| `--continue-on-error` | Keep going across the remaining targets after a failure |

Some workspace-aware commands also expose `--parallel` when parallel fan-out
is a good fit for that command. Treat `--parallel` as command-specific rather
than universal; each command's help text is the source of truth.

These flags are used by commands such as `exec`, `status`, `list`, `add`,
`commit`, `pull`, `push`, `update`, `sync`, `rm`, and `untrack`. Commands keep
their own default root behavior: for example, `status`, `pull`, `push`, `sync`,
and `exec` include the umbrella root by default, while `list` and `update` do
not. `init` and `track` reject these flags, and `untrack` does not support
`--root-only`.

## Usage

### `gyat init`

Initialize a gyat workspace in the current directory, or validate and
reinitialize an existing one.

```sh
gyat init
```

This creates `.gyat` if it does not exist, initializes the root git repository
if needed, and reconciles the gyat-managed `.gitignore` block. Nested gyat
workspaces are rejected.

### `gyat track`

Clone and register a repository in the current workspace.

```sh
# Remote URL
gyat track https://github.com/org/service-auth
gyat track https://github.com/org/service-auth services/auth

# Track a specific branch
gyat track --branch main https://github.com/org/service-auth services/auth

# Local path (relative paths are preferred)
gyat track ../service-auth
gyat track ../service-auth services/auth
```

When tracking a local repository, prefer a relative path such as
`../service-auth` over an absolute path. Relative paths are portable; absolute
paths only work on the machine that recorded them.

After tracking, commit the resulting `.gyat` and `.gitignore` changes in the
umbrella repository.

### `gyat list`

List the repositories tracked in `.gyat` with their path, branch, current
commit, status, and source URL.

```sh
gyat list

# Inspect only the umbrella repository itself
gyat list --root-only
```

### `gyat status`

Show working tree status across the umbrella repository and tracked repos.

In interactive terminals, `gyat status` pages the rendered report
automatically. Use `--no-pager` to print directly to stdout, and note that
redirected or piped output bypasses the pager automatically. Use
`--changed-only` to hide clean repositories and focus on repos that need
attention.

```sh
# Umbrella repository + all tracked repos
gyat status

# Only the auth repo, excluding the umbrella root
gyat status --repo auth --no-root

# Only the umbrella repository
gyat status --root-only

# Print directly without paging
gyat status --no-pager

# Show only repositories with changes or unavailable state
gyat status --changed-only
```

Each target gets its own section that mirrors `git status`: staged changes,
unstaged changes, and untracked files. Tracked repos that are listed in `.gyat`
but missing on disk are shown as `not cloned`.

### `gyat exec`

Run an arbitrary command across selected workspace targets.

```sh
# Run in the umbrella root and every tracked repo
gyat exec -- git status --short

# Run only in backend repos
gyat exec --group backend -- go test ./...

# Run only in the auth repo
gyat exec --repo auth --no-root -- git rev-parse --abbrev-ref HEAD
```

`exec` is the most generic multi-repo primitive in gyat. When the command you
want is not built in, start here. `exec` currently runs targets in resolved
order and prints one output block per target.

### `gyat add`

Stage changes across the workspace.

```sh
# Stage everything in the umbrella root and every cloned tracked repo
gyat add

# Stage a root file
gyat add .gitignore

# Stage all changes inside one tracked repo
gyat add services/auth

# Stage one file inside a tracked repo
gyat add services/auth/handler.go

# Stage the same path inside selected repos
gyat add --repo services/auth --repo services/billing go.mod
```

Without selector flags, `add` routes each path to the repository that owns it.
With selector flags, the supplied path arguments are applied inside each
selected target.

### `gyat commit`

Commit staged changes across selected repos and the umbrella repository with the
same message.

```sh
# Commit staged changes everywhere they exist
gyat commit -m "feat: add login endpoint"

# Commit only selected tracked repos
gyat commit -m "fix: typo" services/auth services/billing

# Commit only the umbrella repository
gyat commit -m "chore: update workspace docs" --root-only

# Skip git hooks
gyat commit -m "wip" --no-verify
```

With no path arguments, gyat commits every tracked repo that currently has
staged changes and then commits the umbrella repository if it also has staged
changes. Root paths or `--root-only` can be used to commit the umbrella
repository by itself. `commit` currently walks selected targets in deterministic
order.

### `gyat pull` and `gyat push`

Pull or push selected workspace targets.

```sh
# Pull everything that has an upstream
gyat pull

# Pull only backend repos, excluding the umbrella root
gyat pull --group backend --no-root

# Push just the auth repo
gyat push --repo auth --no-root
```

Tracked repos that use a local-path remote are skipped with a hint, since there
is no portable remote to pull from or push to. `pull` and `push` currently walk
selected targets in deterministic order.

### `gyat update`

Fast-forward tracked repos to the latest commit on their configured branch.

```sh
# Update all tracked repos
gyat update

# Update one tracked repo
gyat update --repo auth

# Update only the umbrella repository
gyat update --root-only
```

If a repo has `branch` set in `.gyat`, gyat updates it from `origin/<branch>`.
Otherwise it uses the repo's current tracking branch. `update` currently walks
selected targets in deterministic order and excludes the umbrella root unless
you pass `--root-only`.

### `gyat sync`

Sync local clones from the `.gyat` manifest.

```sh
# Reconcile remotes, clone missing repos, and sync the root .gitignore block
gyat sync

# Sync one repo's remote without touching the umbrella root
gyat sync --repo auth --no-root
```

`sync` updates `origin` URLs, clones tracked repos that are missing on disk, and
reconciles the gyat-managed block in the umbrella root `.gitignore`.

### `gyat rm`

Remove files from the working tree and index across the workspace.

```sh
# Remove a root file
gyat rm .gitignore

# Remove a file inside a tracked repo
gyat rm services/auth/handler.go

# Remove a shared file across selected repos
gyat rm --repo services/auth --repo services/billing generated.lock
```

`rm` is workspace-aware, so paths are routed to the repo that owns them unless
selector flags are used. To remove an entire tracked repository from the
workspace, use `gyat untrack` instead.

### `gyat untrack`

Remove one or more tracked repos from the workspace.

```sh
gyat untrack services/auth

# Untrack several repos selected by group
gyat untrack --group experimental
```

This deletes the repo working tree, removes the repo from `.gyat`, and updates
the gyat-managed `.gitignore` block. After untracking, commit the resulting
changes to `.gyat` and `.gitignore` in the umbrella repository.

## Output conventions

gyat follows the same output conventions as git:

| Type | Stream | Format |
|------|--------|--------|
| Data | stdout | plain text or tables suitable for piping |
| Progress | stderr | lowercase, ends with `...` such as `cloning 'services/auth'...` |
| Completion | stderr | lowercase, no trailing punctuation, such as `tracked repository 'services/auth'` |
| Warning | stderr | `warning:` prefix, such as `warning: tracked repository 'services/auth' is not cloned, skipping` |
| Hint | stderr | `hint:` prefix, such as `hint: commit the changes to .gyat and .gitignore` |
| Error | stderr | returned as an error and printed by Cobra |

Progress and hints go to stderr so that data output like `gyat list` can be
piped without noise.

## How it works

gyat is a thin wrapper around the `git` binary. It does not use a Go git
library. The workspace model lives in `.gyat`, and git operations are delegated
to `internal/git`:

- `git.Run(...)` captures stdout and returns `(string, error)`. It is used when output must be parsed, such as `git status`, branch detection, or remote configuration.
- `git.RunInteractive(...)` passes stdin/stdout/stderr through directly for commands that should stream live output, such as `pull`, `push`, and `update`.

Ordered fan-out and optional parallel execution live in `internal/workspace`.
Only commands that opt into `--parallel` expose it.

## License

See [LICENSE.md](LICENSE.md) for details.
