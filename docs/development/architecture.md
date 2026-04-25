# Architecture

smithy-cli is the unified `smithy` CLI for the smithy stack. It runs
single MCP servers and agents, and orchestrates multi-server flows
via `compose` — a daemon-backed supervisor adapted from procsmithy.

## Design Principles

1. **Thin CLI layer** — runtime behaviour lives in the smithy
   runtimes (mcpsmithy, agentsmithy); this binary wires Kong commands
   to them.
2. **Embed, don't fork** — `smithy mcp` embeds the upstream
   `mcpsmithy/pkg/cmd.Commands` rather than reimplementing them.
3. **Everything internal** — command structs and orchestration live
   under `internal/`. No public Go API.

## Package Layout

```
cmd/smithy/main.go              Entry point: kong.Parse() + slog setup
cmd/gen-docs/                   Renders docs/user/reference/cli/ from the Kong tree
internal/commands/              Top-level Kong CLI struct and command groups
internal/commands/compose/      Multi-server `compose` subcommands
```

## Command Surface

```
smithy
├── mcp        Run / manage a single MCP server (embeds mcpsmithy)
├── agent      Run / chat with a single agent
└── compose    Run / manage multi-server flows
    ├── up / down / restart
    ├── ps / status / logs / attach
    └── setup / validate
```
