package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	agentcmd "github.com/iorubs/agentsmithy/pkg/cmd"
	"github.com/iorubs/smithy-cli/internal/agentchat"
	"github.com/iorubs/smithy-cli/internal/commands/stack"
	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
	"github.com/iorubs/smithy-cli/internal/tui"
)

// AgentCmd groups all agentsmithy subcommands.
//
// `Up`/`Down`/`Chat` are smithy-local: they route through the stack
// daemon. The remaining subcommands (`serve`, `validate`, `setup`)
// come from the embedded [agentcmd.Commands], so they share code with
// the standalone `agentsmithy` binary.
type AgentCmd struct {
	Up                AgentUpCmd   `cmd:"" help:"Start a named agent via the stack daemon."`
	Down              AgentDownCmd `cmd:"" help:"Stop a named agent in the stack daemon."`
	Chat              AgentChatCmd `cmd:"" help:"Chat with a daemon-supervised agent over a2a or mcp-http."`
	agentcmd.Commands `embed:""`
}

// AgentUpCmd starts a single named agent through the stack daemon.
// The daemon is spawned if it is not already running.
type AgentUpCmd struct {
	stack.ConfigFlag
	Name   string `arg:"" help:"Agent name (must exist in the stack file)."`
	Detach bool   `help:"Return after the service starts instead of opening the dashboard." short:"d"`
}

// Run executes agent up.
func (m *AgentUpCmd) Run(ctx context.Context) error {
	if err := m.EnsureValid(); err != nil {
		return err
	}
	stackName, err := daemon.DeriveName(m.Config)
	if err != nil {
		return fmt.Errorf("agent up: %w", err)
	}
	paths, err := daemon.PathsFor(stackName)
	if err != nil {
		return fmt.Errorf("agent up: %w", err)
	}

	_, err = daemon.SpawnDetached(ctx, stackName, m.Config, daemon.SpawnTimeout, false)
	switch {
	case errors.Is(err, daemon.ErrAlreadyRunning):
	case errors.Is(err, daemon.ErrNameConflict):
		return fmt.Errorf("agent up: %w", err)
	case err != nil:
		return fmt.Errorf("agent up: spawn daemon: %w", err)
	}

	if err := ipc.NewClient(paths.Socket).StartService(ctx, m.Name, ipc.KindAgent); err != nil {
		return fmt.Errorf("agent up: start %q: %w", m.Name, err)
	}
	if m.Detach {
		fmt.Printf(" Started  %s\n", m.Name)
		return nil
	}
	return tui.Run(stackName, paths.Socket, paths.DaemonLog, false)
}

// AgentDownCmd stops a single named agent in the stack daemon
// without affecting other services.
type AgentDownCmd struct {
	stack.ConfigFlag
	Name string `arg:"" help:"Agent name."`
}

// Run executes agent down.
func (m *AgentDownCmd) Run(ctx context.Context) error {
	stackName, err := daemon.DeriveName(m.Config)
	if err != nil {
		return fmt.Errorf("agent down: %w", err)
	}
	paths, err := daemon.PathsFor(stackName)
	if err != nil {
		return fmt.Errorf("agent down: %w", err)
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
	if err := client.StopService(ctx, m.Name, ipc.KindAgent); err != nil {
		return fmt.Errorf("agent down: stop %q: %w", m.Name, err)
	}
	fmt.Printf(" Stopped  %s\n", m.Name)
	return nil
}

// AgentChatCmd opens a bubbletea chat TUI bound to a daemon-supervised
// agent. The agent must declare transport `a2a` in the stack file;
// other transports are not yet supported by chat.
type AgentChatCmd struct {
	stack.ConfigFlag
	Stack string `help:"Stack name (defaults to the stem of --config)." short:"s"`
	Name  string `arg:"" help:"Agent name (must exist in the stack file)."`
	Reset bool   `help:"Discard any persisted contextID and start a fresh conversation."`
}

// Run resolves the agent against the running stack and starts the chat TUI.
func (c *AgentChatCmd) Run(ctx context.Context) error {
	stackName, err := stack.ResolveStackName(c.Stack, c.Config)
	if err != nil {
		return err
	}
	target, err := agentchat.Resolve(ctx, stackName, c.Name)
	if err != nil {
		return fmt.Errorf("agent chat: %w", err)
	}
	client, err := agentchat.NewClient(ctx, target, c.Reset)
	if err != nil {
		return fmt.Errorf("agent chat: %w", err)
	}
	defer client.Close()
	return tui.RunChat(client, target.Name)
}
