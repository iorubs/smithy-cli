# Testing

Tests use the Go standard `testing` package exclusively; no
third-party test frameworks or assertion libraries.

## Running Tests

```bash
go test ./...              # all packages
go test -cover ./...       # with coverage summary
go test ./internal/tui/... # single package
```

## What to Test

Focus on the public API of the package under test, boundary
conditions, error paths, and parsing logic. Tests exercise the
*module they live in*: do not redundantly verify the behaviour of
the standard library or third-party dependencies.

Specifically:

- Schema parsing and validation (`internal/config/...`)
- Translate round-trips and error cases (`internal/runtime`)
- IPC client/server contracts (`internal/runtime/ipc`)
- Setup tool handlers (`internal/setup`)
- TUI log formatting (`internal/tui`)

Command `Run()` bodies are thin wrappers; test them only when there is
non-trivial orchestration logic that isn't covered by the packages they
call.

## Conventions

- Prefer **table-driven tests** when a function has multiple
  input/output variations. Standalone tests are fine when the setup is
  unique enough that a table adds awkwardness.
- Combine happy-path and error cases in the same table using a
  `wantErr` field when the test structure is identical.
- **Never ignore errors** in test code. Every error from setup
  (`os.WriteFile`, `os.MkdirAll`, `filepath.Abs`, `json.Marshal`,
  constructors, etc.) must be checked with `t.Fatal(err)` or surfaced
  via `http.Error` inside an `httptest` handler. `_ = ...` and
  `x, _ := ...` are forbidden in tests.
- Use `t.Helper()` in assertion helpers so failure lines point to the
  call site, not the helper.
- Keep helpers minimal. A helper that wraps a single stdlib call
  without adding error handling, table semantics, or shared state is
  just noise; inline it.
- **No real network in unit tests.** Anything that needs a remote
  endpoint goes behind `httptest.Server`, a fake, or a build tag
  (`//go:build integration`).
