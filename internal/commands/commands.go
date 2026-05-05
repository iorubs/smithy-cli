// Package commands implements the CLI subcommands.
package commands

import (
	"github.com/iorubs/smithy-cli/internal/commands/stack"
)

// CLI is the root Kong CLI struct.
type CLI struct {
	LogLevel string    `help:"Log level (one of: ${enum})." default:"info" enum:"debug,info,warn,error" short:"l"`
	MCP      MCPCmd    `cmd:"" help:"Run and manage MCP Smithy servers."`
	Agent    AgentCmd  `cmd:"" help:"Run and manage Agent Smithy servers."`
	Stack    stack.Cmd `cmd:"" help:"Run and manage multi Agent and MCP server stacks."`

	Daemon stack.DaemonCmd `cmd:"" name:"__daemon__" hidden:"" help:"Internal: run a stack as a daemon (re-exec target of stack up -d)."`
}
