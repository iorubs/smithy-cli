## mcps Section Guide

The `mcps:` map declares each MCP server in the stack. Use it to
list every backend the agents will eventually talk to. The field
reference below covers every field; this guide covers strategy.

### One Entry Per Server

Each entry is one supervised process. Pick names that read well in
status output and log lines (`smithy stack ps`,
`smithy stack logs <name>`). Keep them short, lowercase, and
specific to the server's job (`fetch`, `docs`, `repo-search`).

### Pointing at Configs

Every entry references a `.mcpsmithy.yaml` that mcpsmithy already
knows how to load. Resolve paths relative to the stack file —
co-locate the configs with the stack file when practical.

### When To Set Transport / Host / Port

Leave the network fields unset for MCPs that only talk to local
agents and let the referenced mcpsmithy config decide. Set them
explicitly when:

- another machine or container needs to reach the server,
- you want a stable address for clients outside the agent fleet, or
- you're running multiple instances of the same MCP and need
  predictable ports.

### Restart Policy

Disable autorestart only for one-shot or hand-managed MCPs that
should not be revived after a clean exit.

### Naming Across `mcps` and `agents`

`mcps` and `agents` are independent namespaces; the same name can
appear in both. Reuse a name only when the MCP and agent are tightly
paired and the shared name aids legibility; otherwise keep them
distinct.
