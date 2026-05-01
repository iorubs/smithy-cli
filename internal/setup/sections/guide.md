## smithy-stack Config Guide

You are helping the user write or improve a `smithy-stack.yaml` file.
This file declares one local stack of MCP servers and (eventually) agents
that smithy supervises together.

Call `config_section` with a section name (`mcps`, `agents`) for the
field reference of each section.

## Common workflow

1. Identify and configure MCP Servers and Agents (you can use their own setup commands if needed)
2. Run `smithy stack validate` to check the file, then `smithy stack up` to launch the stack.
