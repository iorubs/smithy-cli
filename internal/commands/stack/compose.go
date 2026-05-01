// Package stack implements the `smithy stack` subcommands for
// running and managing multi-server stacks of MCP servers and agents.
package stack

// Cmd groups all `smithy stack` subcommands.
type Cmd struct {
	Up       UpCmd       `cmd:"" help:"Start a named stack and follow its log."`
	Ls       LsCmd       `cmd:"" help:"List all stacks under ./.smithy/."`
	Ps       PsCmd       `cmd:"" help:"List services in a running stack."`
	Logs     LogsCmd     `cmd:"" help:"Stream the daemon log for a stack."`
	Down     DownCmd     `cmd:"" help:"Stop a running stack."`
	Validate ValidateCmd `cmd:"" help:"Validate a stack file."`
	Setup    SetupCmd    `cmd:"" help:"Run setup steps for stack services."`
}

// ConfigFlag is the shared `--config` flag embedded in subcommands
// that need to read a stack file (up, validate). Lifecycle
// commands that operate on running stacks (down, logs, ls) address
// stacks by name and do not embed this.
type ConfigFlag struct {
	Config string `help:"Path to config." default:"smithy-stack.yaml" type:"path" short:"c"`
}
