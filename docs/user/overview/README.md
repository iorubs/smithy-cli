---
slug: /
---

# Overview

![smithy](../static/img/body.png)
> **Smithy** is the unified CLI for the smithy stack; it runs single
MCP servers, single agents, and orchestrates multi-server stacks via `smithy stack`.

## Why does it exist?

Engineers are adopting AI coding assistants everywhere, but tuning
their flows is hard and inconsistent. You end up babysitting an agent
instead of training it.

Smithy gives you the building blocks to compose deterministic flows
on top of [mcpsmithy](https://iorubs.github.io/mcpsmithy/) and
[agentsmithy](https://iorubs.github.io/agentsmithy/); declare your
tools, agents, and guardrails in YAML, then iterate until the flow
behaves exactly the way you want.

- **One CLI, the whole stack**; run a single MCP server, a single
  agent, or orchestrate many of them with `smithy stack`.
- **Declarative, not bespoke**; flows live in config, not glue code,
  so they're reviewable, diffable, and reproducible.
- **Tune, don't babysit**; add your own security guards and
  conventions so the agent stays on rails by construction.

## How it works

1. You describe your servers and agents in YAML
   (`.mcpsmithy.yaml`, `.agentsmithy.yaml`, or a `smithy-stack.yaml`
   for a multi-process flow).
2. You run a single component with `smithy mcp serve` or
   `smithy agent serve`, or bring the whole stack up with
   `smithy stack up`; a built-in supervisor that spawns, monitors
   and restarts each process for you.
3. Your MCP-compatible AI assistant connects to the resulting servers
   and uses the tools and agents you defined.

## Config Overview

### Reference docs:
- [Config Reference](../reference/config/README.md)
- [CLI Reference](../reference/cli/README.md)
