package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	mcpcmd "github.com/iorubs/mcpsmithy/pkg/cmd"
	"github.com/iorubs/smithy-cli/internal/commands/stack"
	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
	"github.com/iorubs/smithy-cli/internal/tui"
)

const mcpSpawnTimeout = 10 * time.Second

// MCPCmd groups all mcpsmithy subcommands plus smithy's own per-service
// start/stop which route through the stack daemon.
type MCPCmd struct {
	Up              MCPUpCmd   `cmd:"" help:"Start a named MCP server via the stack daemon."`
	Down            MCPDownCmd `cmd:"" help:"Stop a named MCP server in the stack daemon."`
	mcpcmd.Commands `embed:""`
}

// MCPUpCmd starts a single named MCP server through the stack daemon.
// The daemon is spawned if it is not already running.
type MCPUpCmd struct {
	stack.ConfigFlag
	Name   string `arg:"" help:"MCP server name (must exist in the stack file)."`
	Detach bool   `help:"Return after the service starts instead of following logs." short:"d"`
}

// Run executes mcp up.
func (m *MCPUpCmd) Run(ctx context.Context) error {
	stackName, err := daemon.DeriveName(m.Config)
	if err != nil {
		return fmt.Errorf("mcp up: %w", err)
	}
	paths, err := daemon.PathsFor(stackName)
	if err != nil {
		return fmt.Errorf("mcp up: %w", err)
	}

	pid, err := daemon.SpawnDetached(ctx, stackName, m.Config, mcpSpawnTimeout, false)
	switch {
	case errors.Is(err, daemon.ErrAlreadyRunning):
	case errors.Is(err, daemon.ErrNameConflict):
		return fmt.Errorf("mcp up: %w", err)
	case err != nil:
		return fmt.Errorf("mcp up: spawn daemon: %w", err)
	default:
		_ = pid // daemon started silently
	}

	if err := ipc.NewClient(paths.Socket).StartService(ctx, m.Name, ipc.KindMCP); err != nil {
		return fmt.Errorf("mcp up: start %q: %w", m.Name, err)
	}
	if m.Detach {
		fmt.Printf(" Started  %s\n", m.Name)
		return nil
	}
	return tui.Run(paths.Socket, paths.DaemonLog, false)
}

// MCPDownCmd stops a single named MCP server in the stack daemon without affecting other services.
type MCPDownCmd struct {
	stack.ConfigFlag
	Name string `arg:"" help:"MCP server name."`
}

// Run executes mcp down.
func (m *MCPDownCmd) Run(ctx context.Context) error {
	stackName, err := daemon.DeriveName(m.Config)
	if err != nil {
		return fmt.Errorf("mcp down: %w", err)
	}
	paths, err := daemon.PathsFor(stackName)
	if err != nil {
		return fmt.Errorf("mcp down: %w", err)
	}
	if !daemon.IsStackRunning(ctx, paths, 500*time.Millisecond) {
		return nil
	}
	client := ipc.NewClient(paths.Socket)
	status, err := client.Status(ctx)
	if err == nil {
		for _, svc := range status.Services {
			if svc.Name == m.Name && svc.State == ipc.StateStopped {
				return nil
			}
		}
	}
	if err := client.StopService(ctx, m.Name, ipc.KindMCP); err != nil {
		return fmt.Errorf("mcp down: stop %q: %w", m.Name, err)
	}
	fmt.Printf(" Stopped  %s\n", m.Name)
	return nil
}
