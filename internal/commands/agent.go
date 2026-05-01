package commands

import (
	agentcmd "github.com/iorubs/agentsmithy/pkg/cmd"
	"github.com/iorubs/smithy-cli/internal/commands/stack"
)

// AgentCmd groups all agentsmithy subcommands.
//
// `Up`/`Down` are smithy-local: they route through the stack daemon.
// The remaining subcommands (`serve`, `validate`, `setup`, `chat`)
// come from the embedded [agentcmd.Commands], so they share code with
// the standalone `agentsmithy` binary.
type AgentCmd struct {
	Up                AgentUpCmd   `cmd:"" help:"Start a named agent via the stack daemon."`
	Down              AgentDownCmd `cmd:"" help:"Stop a named agent in the stack daemon."`
	agentcmd.Commands `embed:""`
}

// AgentUpCmd runs a single agentsmithy under the supervisor daemon via the stack file.
type AgentUpCmd struct {
	stack.ConfigFlag
	Name string `arg:"" help:"agent name (must exist in the stack file)"`
}

// Run executes the agent up command.
func (m *AgentUpCmd) Run() error {
	return nil
}

// AgentDownCmd stops a named agent in the stack daemon.
type AgentDownCmd struct {
	stack.ConfigFlag
	Name string `arg:"" help:"agent name (must exist in the stack file)"`
}

// Run executes the agent down command.
func (m *AgentDownCmd) Run() error {
	return nil
}
