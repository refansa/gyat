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

For full command reference and examples, see [USAGE.md](USAGE.md).

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
