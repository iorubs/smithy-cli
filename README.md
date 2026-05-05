# smithy-cli

> The unified `smithy` CLI for the smithy stack. Run a single MCP
> server, run a single agent, or orchestrate multi-server flows
> with `smithy stack`.

![smithy-cli forge](docs/images/forge.png)

**smithy-cli** is the front-end for [mcpsmithy](https://github.com/iorubs/mcpsmithy)
and agentsmithy. It embeds the upstream commands and adds a
daemon-backed `stack` supervisor for running stacks of MCP servers
and agents together.

## Quick Start

```bash
# From source (Go 1.26+)
go build -o bin/smithy ./cmd/smithy

# Single MCP server
smithy mcp --help

# Single agent
smithy agent --help

# Multi-server flow
smithy stack up -c smithy-stack.yaml
```

## Documentation

### For Users

| | |
|---|---|
| [Docs site](https://iorubs.github.io/smithy-cli/) | Documentation overview |
| [CLI Reference](docs/user/reference/cli/README.md) | Generated command reference |

### For Contributors

| | |
|---|---|
| [Development Guide](docs/development/README.md) | Architecture, CLI design, testing |
