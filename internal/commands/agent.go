package commands

// MCPCmd groups all mcpsmithy subcommands.
type AgentCmd struct {
	Up   AgentUpCmd   `cmd:"" help:"Run an Agent Smithy server with the supervisor."`
	Chat AgentChatCmd `cmd:"" help:"Chat with a running agent."`
}

// AgentUpCmd runs a single mcpsmithy under the supervisor daemon via the compose file.
type AgentUpCmd struct {
	ConfigFlag
	Name string `arg:"" help:"agent name (must exist in the compose file)"`
}

func (m *AgentUpCmd) Run() error {
	return nil
}

// AgentChatCmd opens a chat session with an running agent.
type AgentChatCmd struct {
	ConfigFlag
	Name string `arg:"" help:"agent name (must exist in the compose file)"`
}

func (m *AgentChatCmd) Run() error {
	return nil
}
