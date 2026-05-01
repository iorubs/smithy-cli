## agents Section Guide

The `agents:` map declares each agent in the stack. The field
reference below covers every field; this guide covers strategy.

> **Status: agent runtime is WIP.** The schema accepts entries by
> name so stack files can reserve service names today, but the
> supervisor does not yet execute agents. Treat entries as
> placeholders until agentsmithy runtime support lands; the rest of
> this guide describes the shape that's coming.

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

### Naming Across `mcps` and `agents`

`mcps` and `agents` are independent namespaces; the same name can
appear in both. Reuse a name only when the MCP and agent are tightly
paired and the shared name aids legibility; otherwise keep them
distinct.
