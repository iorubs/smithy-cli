# Config

Auto-generated schema and authoring reference for `smithy-stack.yaml`.

## Schema Versions

- [Version 1](v1.md)

---

## What the File Does

`smithy-stack.yaml` declares the MCP servers (and eventually agents)
that `smithy stack` supervises. `smithy stack up` reads this file,
starts each service in-process, and keeps them running. `smithy stack
validate` checks it without starting anything.

## Minimal Working Example

```yaml
version: "1"

mcps:
  docs:
    config: ./docs.mcpsmithy.yaml
    transport: stdio
```

With HTTP servers on fixed ports:

```yaml
version: "1"

mcps:
  api:
    config: ./api.mcpsmithy.yaml
    transport: http
    port: 8081
  search:
    config: ./search.mcpsmithy.yaml
    transport: http
    port: 8082
```

## Config Overview

| Section  | Purpose |
|----------|---------|
| `mcps`   | Map of MCP servers to supervise. Each entry references a `.mcpsmithy.yaml` and declares its transport. |
| `agents` | Map of agent entries (reserved; runtime support lands when agentsmithy is wired in). |

For full field definitions see the version reference above.
Use [`smithy stack setup`](../../getting-started/guided-setup.md) to
author a stack file with an LLM assistant.
