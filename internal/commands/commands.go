// Package commands implements the CLI subcommands.
package commands

import (
	"log/slog"

	"github.com/iorubs/smithy-cli/internal/commands/compose"
)

// LogLevel represents a supported log verbosity level.
type LogLevel string

// Supported log verbosity levels.
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// CLI is the root Kong CLI struct.
type CLI struct {
	LogLevel LogLevel    `help:"Log level (one of: ${enum})." default:"info" enum:"debug,info,warn,error" short:"l"`
	MCP      MCPCmd      `cmd:"" help:"Run and manage MCP Smithy servers."`
	Agent    AgentCmd    `cmd:"" help:"Run and manage Agent Smithy servers."`
	Compose  compose.Cmd `cmd:"" help:"Run and manage multi Agent and MCP server flows."`
}

// ConfigFlag is the shared `--config` flag embedded in subcommands
// that need a compose file path.
type ConfigFlag struct {
	Config string `help:"Path to config." default:"smithy-compose.yaml" type:"path" short:"c"`
}

// ParseLogLevel maps the CLI log-level flag to slog.Level.
func ParseLogLevel(l LogLevel) slog.Level {
	switch l {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
