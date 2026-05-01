---
sidebar_position: 2
---

# Advanced Troubleshoot

> **Status: stub.** This use case lands once agent runtime support is
> wired into smithy. The shape is documented here so the stack
> schema and supervisor can be designed against a real target.

## User story

As an **on-call engineer triaging an incident**, I want one or more
MCP servers exposing telemetry, logs, and runbook docs to a small
fleet of agents; some running in parallel (e.g. log search, metrics
scan) and some in sequence (e.g. summariser → root-cause hypothesiser
→ remediation suggester); so that I get a synthesised view of the
incident without hand-stitching tools.

## Goals

- Multiple MCP servers, each owning a data source.
- Multiple agents, mixing parallel discovery and sequential
  reasoning.
- One supervised stack so a single `smithy stack up` brings the
  whole troubleshooting environment online.

## Technical overview (planned)

- `mcps:` declares each data-source MCP (logs, metrics, runbooks).
- `agents:` declares one orchestrator plus several specialist agents.
  Parallel agents run side-by-side; sequential agents are wired via
  the orchestrator's prompt graph (not via stack ordering).
- Per-agent MCP wiring routes each agent to only the tools it needs.

When agent runtime support lands this page becomes a worked example
with a full stack snippet. Until then, see [Simple
Chat](./simple-chat) for the smallest currently-supported stack.
