---
sidebar_position: 1
---

# Roadmap

Tracked feature ideas, technical debt, and enhancements for the
Smithy CLI.

## Improvements

### Standalone `smithyd` daemon

**Problem:** `smithy stack up -d` puts the stack in the background by re-execing the same `smithy` binary into a hidden internal subcommand. The supervisor that watches MCP servers and agents therefore lives inside the same binary the user runs interactively, which is the wrong shape for a long-running supervisor:

- The supervisor cannot be registered with `systemd`/`launchd` and cannot survive a machine reboot, because its entry point is a hidden subcommand of the user-facing CLI rather than its own program. Operators have to remember to re-run `smithy stack up -d` after every reboot.
- Every chat-only transport that runs the agent over stdio (`mcp-stdio`, `stdio`) needs a supervisor-side relay, because the supervisor already owns the supervised process's stdin/stdout pipes — the CLI cannot share them. A standalone daemon makes that relay a clean IPC contract instead of an in-process detail.

**Value:** A dedicated `smithyd` binary would let `smithy stack up -d` exec `smithyd -f <stack>` instead of re-execing itself, drop the `__daemon__` hidden subcommand and the re-exec plumbing, and unblock chat support for stdio-shaped transports through a daemon-owned IPC relay (`smithy agent chat <name>` → daemon → supervised process). It would also let operators register `smithyd` with the platform service manager so the stack can survive reboots without a login shell. The CLI surface stays the same; only the implementation behind `stack up -d` changes.

**Why parked:** Today's deployments run a single user against a single local stack; the re-exec hack works fine in that environment, and committing to a daemon IPC contract before chat over `mcp-stdio` / `stdio` has a forcing use case risks designing for the wrong shape. Revisit when chat for stdio-shaped agent transports needs to land, or when a deployment needs the supervisor to outlive the login session.

### `smithy chat` with borrowed model

**Problem:** Today an agent can borrow a model from an MCP host (VS Code, Claude Desktop, etc.) through sampling, but only when the host is the one driving the conversation. To talk to the agent the user has to go through the host's chat panel and ask it to forward the question, the host becomes a middleman for a conversation the user could have had directly. The alternative, `smithy chat`, requires the agent to carry its own API key.

**Value:** Letting `smithy chat` use a host-provided model would give the user a direct terminal chat with the agent while keeping model credentials in the host. The agent stays credential-free, and there's no need to round-trip through Copilot or Claude just to ask it something.

A related variant: agents could be started directly by the sampling provider (the MCP host) rather than by `smithy stack up`, and the CLI would still be able to attach to them and chat, even though it doesn't own the supervising process. Same end result for the user, different ownership of the lifecycle.

**Why parked:** Not a priority for the initial MVP release; we just haven't got around to it.
