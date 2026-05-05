# CLI Design

smithy-cli uses [Kong](https://github.com/alecthomas/kong). The Go
struct tree in `internal/commands` is the source of truth; flags,
help, defaults, and enums are declared via struct tags. `gen-docs`
reads the same tree to produce the user-facing reference at
[`docs/user/reference/cli/`](../user/reference/cli/README.md).

Do not document flags or behaviour here; that lives in the generated
reference docs and must stay in sync with the code automatically.

## Layout

```
cmd/smithy/main.go              kong.Parse() + slog setup, no logic
internal/commands/commands.go   Root CLI struct (LogLevel, MCP, Agent, Stack)
internal/commands/{mcp,agent}.go  Single-server subcommands
internal/commands/stack/        Stack subcommands (one file each)
cmd/gen-docs/                   Reference doc generator
```

Each subcommand is a struct with a `Run(...) error` method. Shared
flags live in `ConfigFlag` and are embedded where needed.

## Conventions

- **Kong tags define the surface.** Help text, defaults, enums, and
  short flags belong in the struct tag, not in `Run`.
- **Embed upstream commands** instead of duplicating them; see
  `MCPCmd` embedding `mcpcmd.Commands`.

## Adding a Command

1. Add a struct with `Run(...) error` in the appropriate package
   (`internal/commands` or `internal/commands/stack`).
2. Wire it into the parent struct with `cmd:""` and `help:""` tags.
3. Run `go run ./cmd/gen-docs` to refresh the reference docs.
