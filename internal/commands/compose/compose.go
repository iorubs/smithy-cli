package compose

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
