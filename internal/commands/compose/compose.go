// Package compose implements the `smithy compose` subcommands for
// running and managing multi-server stacks of MCP servers and agents.
package compose

// Cmd groups all `smithy compose` subcommands.
type Cmd struct {
	Up       UpCmd       `cmd:"" help:"Start all Smithy Agent and MCP servers."`
	Ps       PsCmd       `cmd:"" help:""`
	Status   StatusCmd   `cmd:"" help:""`
	Logs     LogsCmd     `cmd:"" help:""`
	Down     DownCmd     `cmd:"" help:""`
	Attach   AttachCmd   `cmd:"" help:""`
	Restart  RestartCmd  `cmd:"" help:""`
	Validate ValidateCmd `cmd:"" help:""`
	Setup    SetupCmd    `cmd:"" help:""`
}
