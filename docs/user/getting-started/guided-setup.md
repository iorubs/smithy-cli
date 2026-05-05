---
sidebar_position: 3
---

# Guided Setup

Three tools work together to author a full smithy stack:
`smithy mcp setup` authors each MCP config, `smithy agent setup`
authors each agent config, and `smithy stack setup` authors the
`smithy-stack.yaml` that ties them together.

## Step 1: Author each MCP config

`smithy mcp setup` starts an MCP server that exposes two tools designed
for config-authoring sessions:

- **`config_guide`**; returns an overview of the `.mcpsmithy.yaml`
  structure and an annotated minimal example.
- **`config_section`**; returns a deep reference for one config section:
  all fields, types, defaults, and valid values.

For each MCP server in your stack:

1. Run `smithy mcp setup` in the directory where the MCP config will live.
2. Connect your MCP-compatible agent (VS Code, Cursor, Claude, etc.).
3. Prompt it to write the config.
4. Validate: `smithy mcp validate`

Repeat for each MCP.

## Step 2: Author each agent config

`smithy agent setup` starts an MCP server that exposes the same two
tools (`config_guide`, `config_section`) for `.agentsmithy.yaml`
instead of `.mcpsmithy.yaml`.

For each agent in your stack:

1. Run `smithy agent setup` in the directory where the agent config
   will live.
2. Connect your agent and prompt it to write the config, referencing
   the MCP endpoints from Step 1 if the agent calls those tools.
3. Validate: `smithy agent validate`

Repeat for each agent.

## Step 3: Author the stack file

`smithy stack setup` starts an MCP server that exposes two tools designed
for stack-authoring sessions:

- **`config_guide`**; returns an overview of the `smithy-stack.yaml`
  structure: top-level keys, how `mcps` and `agents` are declared, and an
  annotated minimal example.
- **`config_section`**; returns a deep reference for one config section
  (`mcps`, `agents`): all fields, types, defaults, and valid values.

Both tools are generated from the same schema that drives
`smithy stack validate`; they are always accurate for the installed
binary version.

1. Run `smithy stack setup` in your project root.
2. Connect your agent and prompt it to write `smithy-stack.yaml`,
   referencing the MCP configs from Step 1.
3. Validate: `smithy stack validate`

## Step 4: Bring it up

```
smithy stack up
```

## Notes

- The agent writes files using its own file tools; smithy stays
  read-only throughout.
- Setup tools are only available in setup mode. They are not exposed by
  `smithy mcp serve`, `smithy agent serve`, `smithy stack up`, or any
  other subcommand.
- If you already have a config and want to improve a specific section,
  skip `config_guide` and call `config_section` directly.
