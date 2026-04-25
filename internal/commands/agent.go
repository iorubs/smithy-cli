package commands

// AgentCmd groups all agentsmithy subcommands.
type AgentCmd struct {
	Up   AgentUpCmd   `cmd:"" help:"Run an Agent Smithy server with the supervisor."`
	Chat AgentChatCmd `cmd:"" help:"Chat with a running agent."`
}

// AgentUpCmd runs a single agentsmithy under the supervisor daemon via the compose file.
type AgentUpCmd struct {
	ConfigFlag
	Name string `arg:"" help:"agent name (must exist in the compose file)"`
}

// Run executes the agent up command.
func (m *AgentUpCmd) Run() error {
	return nil
}

// AgentChatCmd opens a chat session with a running agent.
type AgentChatCmd struct {
	ConfigFlag
	Name string `arg:"" help:"agent name (must exist in the compose file)"`
}

// Run executes the agent chat command.
func (m *AgentChatCmd) Run() error {
	return nil
}
