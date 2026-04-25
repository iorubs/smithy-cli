# Testing

Standard `testing` package only — no third-party test frameworks or
assertion libraries.

```bash
go test ./...
go test -cover ./...
```

## Conventions

- Prefer **table-driven tests** when a function has multiple
  input/output variations.
- Combine happy-path and error cases in the same table using a
  `wantErr` field when the test structure is identical.
- Focus on command wiring, flag parsing, and orchestration logic.
  Runtime behaviour is tested in the upstream runtime repos.
