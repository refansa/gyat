# AGENTS.md

## Go Project Setup

- **Go version**: 1.26+ required (see `go.mod`)
- **Build**: `go build -o bin/gyat.exe .`
- **Run**: `go run .` or `./bin/gyat.exe` (pre-built binary exists)
- **Test**: `go test ./...` (no tests currently)

## Project Structure

- `cmd/` - Cobra commands
- `internal/` - Core packages: `git`, `manifest`, `pager`, `workspace`, `runtime`
- `main.go` - Entry point

## Key Commands

```sh
gyat --help          # Show all commands
gyat init            # Initialize workspace
gyat track <path>    # Track a child repo
gyat exec -- git status  # Run git across all repos
```

## Notes

- This is a thin wrapper around `git` binary (no go-git library)
- Uses spf13/cobra for CLI
- No existing tests - agent should add tests when modifying code
- Module path: `github.com/refansa/gyat/v2`

## Testing

- Test umbrella workspace: `pwsh scripts/init-test.ps1` (Windows) or `bash scripts/init-test.sh` (Linux/macOS)
- Test workspace location: `tmp/gyat-test/` (already in .gitignore)