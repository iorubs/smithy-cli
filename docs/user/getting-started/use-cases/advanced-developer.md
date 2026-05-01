---
sidebar_position: 3
---

# Advanced Developer

> **Status: stub.** This use case lands once agent runtime support is
> wired into smithy. The shape is documented here so the stack
> schema and supervisor can be designed against a real target.

## User story

As an **engineer building a feature with AI assistance**, I want a
team of agents; orchestrator, researcher, reviewer (each with their
own sub-agents), and implementer; running over a shared set of MCP
servers so that planning, research, code review, and implementation
each happen in their own context with the right tools, models, and
skills.

## Goals

- One stack supervises the full agent team.
- Each agent has its own MCP tool subset, model, and skill profile.
- Sub-agents are first-class citizens, not prompt-only fan-out.

## Technical overview (planned)

- `mcps:` declares the shared tool servers (code search, repo ops,
  docs, build/test, etc.).
- `agents:` declares the orchestrator plus the researcher, reviewer,
  and implementer roles. Each agent points at its own
  `.agentsmithy.yaml` so model, system prompt, and skills are
  per-agent.
- Sub-agents (e.g. researcher's sub-researchers) are declared as
  their own entries; the orchestrator's prompt graph delegates to
  them.

When agent runtime support lands this page becomes a worked example
with a full stack snippet showing per-agent MCP wiring and
delegation. Until then, see [Simple Chat](./simple-chat) for the
smallest currently-supported stack.
