---
sidebar_position: 1
---

# Simple Chat

## User story

As an **engineer experimenting with a single-purpose AI assistant**,
I want one MCP server providing tools to one autonomous agent so that
I can iterate on prompts and tools without wiring up infrastructure.

## Goals

- One MCP server delivers domain tools (search, fetch, etc.).
- One agent connects to that MCP and handles user chat.
- One stack file describes the whole stack; one command runs it.

## Technical overview

**Topology:** one MCP, one agent, both supervised by `smithy stack up`.

**Workflow:**

1. Author the MCP config with `mcpsmithy setup`.
2. Author the stack file with `smithy stack setup` (or by hand).
3. `smithy stack validate` to check.
4. `smithy stack up` to run the stack in the foreground.

For full field reference, see the [config
reference](../../reference/config/README.md).

:::tip Generate this config with your agent
Run `smithy stack setup` and prompt:

> Set up a simple chat stack with one MCP server named `fetch` that
> serves `./fetch.mcpsmithy.yaml` over HTTP on port 8081, and one
> agent named `assistant`.

See [Guided Setup](../guided-setup.md) for the full workflow.
:::
