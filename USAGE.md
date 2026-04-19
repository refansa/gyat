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
