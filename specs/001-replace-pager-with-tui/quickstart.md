Quickstart: Using the new TUI viewer (developer notes)

1. Build the binary:

   go build ./...

2. Interactive usage (TTY):

    Run a command that produces multi-screen output in a terminal. The TUI viewer should open automatically when stdin/stdout are TTYs and paging is enabled.

    Smoke helper:

    go run ./tools/pager-smoke -n 10000

    Manual verification:

    - `j` / `k` moves one line
    - `space` / `b` moves one page
    - `/` opens search; `enter` runs it
    - `n` / `N` jumps between matches
    - `?` or `h` toggles help
    - `q` exits immediately back to the shell

3. Non-interactive / piped usage:

   Pipe output to a file or another process: the pager will not be invoked, and raw output will be written unchanged. Example:

   gyat status | tee out.txt

4. Disable pager explicitly:

   - Use the flag: --no-pager (when available on the specific command)
   - Or set environment variable: GYAT_NO_PAGER=1

5. Developer testing:

    - Unit tests: run `go test ./internal/pager` to run pager package tests
    - Integration tests: run `go test ./tests/...` and ensure piped vs TTY behaviors are covered
    - Formatting: run `gofmt -w ./cmd ./internal ./tests ./tools`
    - Vet: run `go vet ./...`

6. Large-output validation:

    Use `go run ./tools/pager-smoke -n 50000` to generate a large output sample for manual performance checks.
