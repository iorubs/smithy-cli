# Server

## smithy

CLI for running and managing smithy MCP Servers and Agents

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

Run a MCP Smithy server with the supervisor.

```
smithy mcp up [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | mcp server name (must exist in the compose file) |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-compose.yaml` | Path to config. |


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

Run an Agent Smithy server with the supervisor.

```
smithy agent up [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | agent name (must exist in the compose file) |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-compose.yaml` | Path to config. |


#### agent chat

Chat with a running agent.

```
smithy agent chat [flags]
```

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | yes | — | agent name (must exist in the compose file) |


| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | `string` | `smithy-compose.yaml` | Path to config. |


### compose

Run and manage multi Agent and MCP server flows.

```
smithy compose [flags]
```

#### compose up

Start all Smithy Agent and MCP servers.

```
smithy compose up [flags]
```

#### compose ps



```
smithy compose ps [flags]
```

#### compose status



```
smithy compose status [flags]
```

#### compose logs



```
smithy compose logs [flags]
```

#### compose down



```
smithy compose down [flags]
```

#### compose attach



```
smithy compose attach [flags]
```

#### compose restart



```
smithy compose restart [flags]
```

#### compose validate



```
smithy compose validate [flags]
```

#### compose setup



```
smithy compose setup [flags]
```
