// Package stack implements the `smithy stack` subcommands for
// running and managing multi-server stacks of MCP servers and agents.
package stack

import (
	"errors"
	"fmt"
	"os"

	"github.com/iorubs/smithy-cli/internal/config"
	"github.com/iorubs/smithy-cli/internal/runtime"
)

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
	Config string `help:"Path to config." default:"smithy-stack.yaml" short:"c"`
}

// EnsureExists returns a friendly error when the configured stack
// file is missing. It is intended to run in the parent process before
// spawning a detached daemon, so the user sees the real cause instead
// of a socket-readiness timeout.
func (c ConfigFlag) EnsureExists() error {
	if _, err := os.Stat(c.Config); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s not found", c.Config)
		}
		return fmt.Errorf("%s: %w", c.Config, err)
	}
	return nil
}

// EnsureValid parses the stack file and translates it so that
// referenced MCP and agent config paths are existence-checked in the
// parent process. Surfaces the real cause instead of a daemon
// socket-readiness timeout.
func (c ConfigFlag) EnsureValid() error {
	if err := c.EnsureExists(); err != nil {
		return err
	}
	data, err := os.ReadFile(c.Config)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s not found", c.Config)
		}
		return fmt.Errorf("%s: %w", c.Config, err)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		return err
	}
	if _, err := runtime.Translate(cfg, c.Config); err != nil {
		return err
	}
	return nil
}
