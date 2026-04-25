# Development Guide

smithy-cli is the CLI front-end for running and managing smithy MCP
servers and agents. The binary is `smithy`. See the project
[README](../../README.md) for user-facing documentation.

## Development Docs

| Document | Scope |
|----------|-------|
| [Architecture](architecture.md) | Package layout and command surface |
| [CLI Design](cli.md) | Kong-based CLI and gen-docs |
| [Testing](testing.md) | Testing conventions |

## Dependencies

stdlib-first. External modules must be justified by significant value
over a stdlib solution.

- `github.com/alecthomas/kong` — declarative CLI parsing.
- `github.com/iorubs/mcpsmithy` — embedded for `smithy mcp` subcommands.

## Conventions

- **Go naming** — `MixedCaps` exports, `mixedCaps` unexported,
  `snake_case.go` filenames, consistent acronym casing (`ID`, `URL`,
  `HTTP`).
- **Comments** — Package comments on primary files; document exported
  symbols with `// Name ...`; comment non-obvious logic only.
- **stderr-only logging** — All output via `slog` to stderr. Stdout is
  reserved for protocol traffic from embedded servers.
