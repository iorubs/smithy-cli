package commands

import (
	mcpcmd "github.com/iorubs/mcpsmithy/pkg/cmd"
)

// MCPCmd groups all mcpsmithy subcommands.
type MCPCmd struct {
	Up              MCPUpCmd `cmd:"" help:"Run a MCP Smithy server with the supervisor."`
	mcpcmd.Commands `embed:""`
}

// MCPUpCmd runs a single mcpsmithy under the supervisor daemon via the compose file.
type MCPUpCmd struct {
	ConfigFlag
	Name string `arg:"" help:"mcp server name (must exist in the compose file)"`
}

// Run executes the mcp up command.
func (m *MCPUpCmd) Run() error {
	return nil
}
