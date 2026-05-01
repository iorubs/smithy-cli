# Server

## smithy

CLI for running and managing smithy MCP Servers and Agents.

```
smithy <command> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-h, --help` | `bool` | — | Show context-sensitive help. |
| `-l, --log-level` | `enum(debug,info,warn,error)` | `info` | Log level (one of: debug,info,warn,error). |


### mcp

Run and manage MCP Smithy servers.

```
smithy mcp [flags]
```

#### mcp up

Start a named MCP server via the stack daemon.

```
smithy mcp up [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | MCP server name (must exist in the stack file). |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |
| `-d, --detach` | `bool` | — | Return after the service starts instead of following logs. |


#### mcp down

Stop a named MCP server in the stack daemon.

```
smithy mcp down [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | MCP server name. |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |


#### mcp serve

Run MCP server.

```
smithy mcp serve [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `.mcpsmithy.yaml` | Path to config. |
| `--transport` | `enum(stdio,http)` | `stdio` | Transport to use. |
| `--addr` | `string` | `:8080` | Listen address (HTTP transport only). |
| `--watch` | `bool` | `false` | Watch config file and hot-reload on change. |


#### mcp validate

Validate config file.

```
smithy mcp validate [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `.mcpsmithy.yaml` | Path to config. |


#### mcp sources

Manage sources.

```
smithy mcp sources [flags]
```

#### mcp setup

Start config-authoring MCP server assistant.

```
smithy mcp setup [flags]
```

### agent

Run and manage Agent Smithy servers.

```
smithy agent [flags]
```

#### agent up

Start a named agent via the stack daemon.

```
smithy agent up [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | agent name (must exist in the stack file) |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |


#### agent down

Stop a named agent in the stack daemon.

```
smithy agent down [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | agent name (must exist in the stack file) |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |


#### agent serve

Start the agent server.

```
smithy agent serve [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `.agentsmithy.yaml` | Path to config. |
| `--transport` | `enum(a2a,stdio,mcp-stdio,mcp-http)` | `a2a` | Transport to use. |
| `--addr` | `string` | `:8080` | Listen address (HTTP-like transports). |
| `--watch` | `bool` | `false` | Watch config file and hot-reload on change. |


#### agent validate

Validate config file.

```
smithy agent validate [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `.agentsmithy.yaml` | Path to config. |


#### agent setup

Start the config-authoring MCP assistant.

```
smithy agent setup [flags]
```

#### agent chat

Chat with the configured agent (minimal stdio REPL).

```
smithy agent chat [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `.agentsmithy.yaml` | Path to config. |
| `-o, --once` | `string` | — | Single-shot input; print response and exit. |
| `-v, --verbose` | `bool` | — | Print tool calls and intermediate steps. |


### stack

Run and manage multi Agent and MCP server stacks.

```
smithy stack [flags]
```

#### stack up

Start a named stack and follow its log.

```
smithy stack up [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | no | — | Stack name. Defaults to the stack file's basename without extension. |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |
| `-d, --detach` | `bool` | — | Return after the daemon is ready instead of following its log. |


#### stack ls

List all stacks under ./.smithy/.

```
smithy stack ls [flags]
```

#### stack ps

List services in a running stack.

```
smithy stack ps [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | no | — | Stack name. |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |


#### stack logs

Stream the daemon log for a stack.

```
smithy stack logs [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | no | — | Stack name. |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |
| `-f, --follow` | `bool` | — | Follow log output in --json mode. |
| `--json` | `bool` | — | Output raw log instead of the interactive TUI. |


#### stack down

Stop a running stack.

```
smithy stack down [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | no | — | Stack name. |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |


#### stack validate

Validate a stack file.

```
smithy stack validate [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |


#### stack setup

Run setup steps for stack services.

```
smithy stack setup [flags]
```

### __daemon__

Internal: run a stack as a daemon (re-exec target of stack up -d).

```
smithy __daemon__ [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-stack.yaml` | Path to config. |
| `--name` | `string` | — | Stack name. |
| `--start-all` | `bool` | — | Start all services on daemon startup. |

