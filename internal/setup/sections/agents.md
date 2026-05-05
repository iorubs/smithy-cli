## agents Section Guide

The `agents:` map declares each agent in the stack. The field
reference below covers every field; this guide covers strategy.

### One Entry Per Agent

Each entry is one supervised agent process. Names show up in
`smithy stack ps`, `smithy stack logs <name>`, and chat
attachments; pick something short and role-specific (`assistant`,
`reviewer`, `orchestrator`).

### Roles, Not Pipelines

The stack file declares **what runs**, not **how agents call each other**.
Sequential or hierarchical orchestration (researcher → reviewer →
implementer, sub-agents under a parent, etc.) lives inside the
agent's prompt graph, not in the stack file. Use one entry per agent
process; including sub-agents that need their own model, tool set,
or skills.

### MCP Wiring

When an agent depends on specific MCPs in this stack file, the
recommendation is to keep the wiring inside the agent's
`.agentsmithy.yaml` (point its tool endpoints at the MCP addresses
declared above). Compose-level wiring fields (e.g. an `mcp:` list)
will land here when the runtime needs them.

### Picking a Transport for `smithy agent chat`

The agent's `transport:` controls how the supervised process
exposes itself; `smithy agent chat <name>` can attach to two of
those transports today:

- `a2a` (default) — full chat experience: persistent conversation
  context across CLI runs, history replay when you reattach, and
  the standard agent-to-agent JSON-RPC contract.
- `mcp-http` — chat is wired as a single tool call per turn. No
  history replay (each turn is a fresh agent session under the
  hood). Pick this when the agent is primarily consumed by other
  MCP clients and chat is just a smoke-test affordance.

`mcp-stdio` and `stdio` cannot be chatted with from `smithy agent
chat` yet — the supervisor owns those processes' stdio pipes, so a
supervisor-side relay is needed first. Until that lands, run those
transports directly with `smithy agent serve --transport stdio` for
interactive use, or expose the agent over `a2a` / `mcp-http` if
you want to chat through the supervised stack.

### Naming Across `mcps` and `agents`

`mcps` and `agents` are independent namespaces; the same name can
appear in both. Reuse a name only when the MCP and agent are tightly
paired and the shared name aids legibility; otherwise keep them
distinct.
