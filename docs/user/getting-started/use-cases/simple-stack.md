---
sidebar_position: 1
---

# Simple Stack

## User story

As an **engineer running an AI workload locally**, I want one stack
file that supervises the MCP servers and agents I need — pulling
tools, prompts, and models together — so that one command brings
the whole environment up and another tears it down, without me
managing each process by hand.

## Goals

- One stack file describes every MCP server and agent.
- One command (`smithy stack up`) starts the daemon and supervises
  the lot; one command (`smithy stack down`) stops everything.
- Per-service configs (`*.mcpsmithy.yaml`, `*.agentsmithy.yaml`)
  stay where they are; the stack file just points at them.
- Logs, status, and chat all flow through `smithy` — no need to
  attach to individual processes.

## Technical overview

**Topology:** any number of `mcps:` and any number of `agents:` in a
single stack. The smithy daemon supervises both kinds; restart
policy, env-file loading, and logging work the same way for either.

**Workflow:**

1. Author each MCP config with `mcpsmithy setup` (or by hand).
2. Author each agent config with `agentsmithy setup` (or by hand).
3. Author the stack file with `smithy stack setup` (or by hand) —
   it lists each service with its config path, transport, and
   address.
4. `smithy stack validate` to check before starting.
5. `smithy stack up` to bring everything online.
6. `smithy agent chat <name>` to talk to an agent. Chat works for
   agents whose `transport:` is `a2a` (default; full history) or
   `mcp-http` (single tool call per turn, no history replay);
   `mcp-stdio` and `stdio` agents need to be driven directly with
   `smithy agent serve` until supervisor-side relay lands.
7. `smithy stack down` to stop everything.

**Scaling up:** the same stack file scales from "one MCP, one agent"
to multi-agent workloads with shared MCPs. Per-agent transports,
addresses, and env files keep the services isolated; the stack file
remains the single source of truth.

For full field reference, see the [config
reference](../../reference/config/README.md).

:::tip Generate this config with your agent
Run `smithy stack setup` and describe the topology. Before
prompting, have ready:

- The MCP servers you want to run (config path, transport, port)
- The agents that should connect to them (config path, transport)
- Any env files the daemon should load before starting services

Then use a prompt like:

> Set up a stack with one MCP server `fetch` serving
> `./fetch.mcpsmithy.yaml` over HTTP on port 8081, and one agent
> `assistant` serving `./assistant.agentsmithy.yaml` over A2A on
> port 9090. Load `.env` automatically.

See [Guided Setup](../guided-setup.md) for the full workflow.
:::
