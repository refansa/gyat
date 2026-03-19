# gyat - Git Your Ass Together

A git submodule manager that aggregates all of your related repositories into one
single umbrella repository, making them easy to manage without wrestling with raw
`git submodule` commands.

## Why?

Sometimes what's sold as "microservices" is really just a **distributed monolith** —
10+ repos that are tightly coupled, deployed together, and should have been a single
repository from the start.

If that sounds familiar, `gyat` is for you.

Managing a constellation of tightly-coupled repos means:

- Jumping between repos just to make one feature work
- Keeping versions in sync manually
- PRs scattered across multiple repos for a single logical change
- Onboarding new devs is a nightmare

`gyat` gives you a single umbrella repository that holds all of these repos as
submodules, with a simple CLI to manage them — without wrestling with raw git
submodule commands.

### I meant, what with the funny name?

Ah, of course, that is the most important bit. Initially I was thinking about naming
the project as "Yet Another Git Tracker" (yagt). But it didn't roll well off the tongue
and it felt like a little bit more awkward to type as a command. Upon a closer look
at the abbreviation, I could rearrange it to make a funny little gag name "gyat" 
(minus the other 't'). It just so happened that I could easily come up with the project's
full name immediately "Git Your Ass Together" because I always thought that these 
repositories are really a big pain in the ass. Why can't they just be bundled
together from the start?!

## Goals

- **Aggregate** multiple repositories under one roof
- **Simplify** common submodule operations (track, add, commit, remove, update, sync, list)
- **Stay out of the way** — it is a thin layer on top of git, not a replacement
- **Make it easy** to add or remove repositories as the project evolves

## Installation

**Using `go install` (recommended, requires Go 1.26+):**

```sh
go install github.com/refansa/gyat@latest
```

The binary will be placed in `$GOPATH/bin` — make sure that directory is on your `PATH`.

**From source (requires Go 1.26+):**

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

**Verify:**

```sh
gyat --help
```

## Usage

### `gyat init`

Initialize a new gyat-managed repository in the current directory, or reinitialize
an existing one. If a `.gitmodules` file is already present (e.g. after cloning an
existing gyat-managed repo), all submodules will be initialized and checked out
automatically.

```sh
gyat init
```

### `gyat track`

Register a repository as a submodule. Accepts both remote URLs and local paths.

```sh
# Remote URL
gyat track https://github.com/org/service-auth
gyat track https://github.com/org/service-auth services/auth

# Track a specific branch
gyat track --branch main https://github.com/org/service-auth services/auth

# Local path (relative — portable across machines)
gyat track ../service-auth
gyat track ../service-auth services/auth
```

When using a local path, prefer a **relative** path (e.g. `../service-auth`) over an
absolute one. Absolute paths are machine-specific and will break for anyone else who
clones the umbrella repository.

### `gyat add`

Stage changes in one or all submodules. Mirrors `git add` but operates across
every registered submodule at once.

```sh
# Stage all changes in every submodule
gyat add

# Stage changes in a specific submodule
gyat add services/auth

# Stage changes in multiple specific submodules
gyat add services/auth services/billing
```

When no path is given, `gyat add` runs `git add -A` inside every submodule that
is currently checked out, leaving clean or uninitialized submodules untouched.

### `gyat status`

Show the working tree status of the umbrella repository and all registered
submodules. Each repository gets its own clearly labelled section that mirrors
`git status`: staged changes, unstaged changes, and untracked files.

```sh
# Show status for all repositories
gyat status

# Show status for specific submodules (plus the umbrella)
gyat status services/auth services/billing
```

Example output:

```
umbrella repository — main
──────────────────────────
Changes to be committed:
	new file:    services/auth

services/auth — feat/login
──────────────────────────
Changes to be committed:
	new file:    handler.go

Changes not staged for commit:
	modified:    main.go

services/billing — main
───────────────────────
nothing to commit, working tree clean
```

Submodules registered in `.gitmodules` but not yet initialised on disk are
flagged with `not initialized` in their section. Pass one or more submodule
paths to limit the output to those submodules — the umbrella is always shown.

### `gyat commit`

Commit staged changes across multiple submodules simultaneously with the same
commit message, then record the updated submodule refs in the umbrella repository.

```sh
# Commit all submodules with staged changes, then the umbrella
gyat commit -m "feat: add login endpoint"

# Commit only specific submodules
gyat commit -m "fix: typo" services/auth services/billing

# Skip git hooks
gyat commit -m "wip" --no-verify
```

With no path arguments, every checked-out submodule that has staged changes is
committed, the updated submodule references are staged in the umbrella repository,
and the umbrella itself is committed — all with the same message.

With one or more path arguments, only the specified submodules are committed. The
umbrella repository is still committed afterwards to record the new submodule SHAs.

| Flag              | Description                              |
|-------------------|------------------------------------------|
| `-m, --message`   | Commit message (required)                |
| `--no-verify`     | Bypass pre-commit and commit-msg hooks   |

### `gyat list`

List all managed submodules with their path, tracked branch, current commit, status,
and URL.

```sh
gyat list
```

Example output:

```
PATH              BRANCH      COMMIT      STATUS        URL
----------------------------------------------------------------------------------------------------
services/auth     main        a1b2c3d4    up to date    https://github.com/org/service-auth
services/billing  (default)   e5f6a7b8    modified      https://github.com/org/service-billing
services/notify   (default)   ?           not initialized  ../service-notify
```

### `gyat untrack`

Untrack a submodule cleanly. This performs the full three-step cleanup that git
requires: deinit, delete the cached module data, and remove from the index.

```sh
gyat untrack services/auth
```

After untracking, commit the resulting changes to `.gitmodules` and the index:

```sh
gyat commit -m "chore: untrack auth submodule"
```

### `gyat update`

Update one or all submodules to the latest commit on their tracked remote branch.

```sh
# Update all submodules
gyat update

# Update a specific submodule
gyat update services/auth
```

### `gyat sync`

Sync each submodule's remote URL from `.gitmodules`. Useful when a repository has
been moved or renamed and all local clones need to point to the new location.

```sh
gyat sync
```

After syncing URLs, any submodules that were not yet cloned are initialized
automatically.

## Output conventions

gyat follows the same output conventions as git:

| Type        | Stream | Format                                                              |
|-------------|--------|---------------------------------------------------------------------|
| Data        | stdout | plain text or table (suitable for piping)                           |
| Progress    | stderr | lowercase, ends with `...` — e.g. `syncing submodule URLs...`      |
| Completion  | stderr | lowercase, no trailing punctuation — e.g. `removed submodule 'x'`  |
| Warning     | stderr | `warning:` prefix — e.g. `warning: 'path' is an absolute path`     |
| Hint        | stderr | `hint:` prefix — e.g. `hint: use a relative path for portability`  |
| Error       | stderr | `error:` prefix (handled by cobra)                                  |

Progress and hints are written to stderr so that data output (e.g. `gyat list`) can
be piped cleanly without noise.

## How it works

gyat is a thin wrapper around the `git` binary. It does not use any Go git library —
every operation shells out to `git` directly. The only abstraction is in
`internal/git`:

- `git.Run(args...)` — captures stdout, returns `(string, error)`. Used when output
  needs to be parsed (e.g. reading `.gitmodules`, parsing submodule status).
- `git.RunInteractive(args...)` — passes stdin/stdout/stderr straight through to the
  terminal. Used for commands that show live progress or prompt for credentials.

## License

See [LICENSE.md](LICENSE.md) for details.
