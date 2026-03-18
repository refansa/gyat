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

## Goals

- **Aggregate** multiple repositories under one roof
- **Simplify** common submodule operations (add, remove, update, sync)
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

### `gyat add`

Add a repository as a submodule. Accepts both remote URLs and local paths.

```sh
# Remote URL
gyat add https://github.com/org/service-auth
gyat add https://github.com/org/service-auth services/auth

# Track a specific branch
gyat add --branch main https://github.com/org/service-auth services/auth

# Local path (relative — portable across machines)
gyat add ../service-auth
gyat add ../service-auth services/auth
```

When using a local path, prefer a **relative** path (e.g. `../service-auth`) over an
absolute one. Absolute paths are machine-specific and will break for anyone else who
clones the umbrella repository.

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

### `gyat remove`

Remove a submodule cleanly. This performs the full three-step removal that git
requires: deinit, delete the cached module data, and remove from the index.

```sh
gyat remove services/auth

# Alias
gyat rm services/auth
```

After removal, commit the resulting changes to `.gitmodules` and the index.

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