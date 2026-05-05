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

- `github.com/alecthomas/kong`: declarative CLI parsing.
- `github.com/iorubs/mcpsmithy`: embedded for `smithy mcp` subcommands; also the MCP runtime used by `smithy stack up`.
- `github.com/charmbracelet/bubbletea` + `bubbles` + `lipgloss`; TUI for `stack logs` / `stack up` dashboard.
- `github.com/modelcontextprotocol/go-sdk`: MCP server for `stack setup` (stdio transport).
- `go.yaml.in/yaml/v4`: YAML config parsing.

## Conventions

- **Go naming.** `MixedCaps` exports, `mixedCaps` unexported,
  `snake_case.go` filenames, consistent acronym casing (`ID`, `URL`,
  `HTTP`).
- **Comments.** Package comments on primary files; document exported
  symbols with `// Name ...`; comment non-obvious logic only.
- **stderr-only logging.** All output via `slog` to stderr. Stdout is
  reserved for protocol traffic from embedded servers.

## Log Levels

The `--log-level` (`-l`) flag sets the minimum level. Default is `info`.

| Level | Use for |
|-------|---------|
| **Error** | Failures that stop the current operation. |
| **Warn** | Unexpected but recoverable conditions; operation continues. |
| **Info** | Significant lifecycle events; server start/stop, daemon ready, service state changes. |
| **Debug** | Per-item detail, internal decisions, IPC wire traffic. |

Decision rules:
1. Network or I/O action starting → Info.
2. Per-item progress within a batch → Debug (summary at the end → Info).
3. Something failed but we continue → Warn.
4. Something failed and we stop → Error.

## Error Messages

- Start with lowercase (Go convention).
- Lead with the operation, then the cause: `stack: read file: …`
- Be specific about what went wrong; don't prescribe a fix.
- Wrap only when it adds context; redundant wrapping adds noise.
