package stack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
	"github.com/iorubs/smithy-cli/internal/tui"
)

// spawnTimeout bounds how long the parent waits for the daemon's
// socket to appear after `stack up` or `mcp up`.
const spawnTimeout = 10 * time.Second

// UpCmd starts a stack as a daemon. Without -d, it follows the
// daemon log until ctrl-C; the daemon keeps running. With -d, it
// returns once the daemon's socket accepts connections.
//
// The stack name defaults to the stack file's basename without
// extension. Pass a positional name to override; this also lets you
// run multiple instances of the same stack file concurrently.
type UpCmd struct {
	ConfigFlag
	Name   string `arg:"" optional:"" help:"Stack name. Defaults to the stack file's basename without extension."`
	Detach bool   `help:"Return after the daemon is ready instead of following its log." short:"d"`
}

// Run executes the up command.
func (c *UpCmd) Run(ctx context.Context) error {
	if err := c.EnsureValid(); err != nil {
		return err
	}
	name, err := ResolveStackName(c.Name, c.Config)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}

	_, err = daemon.SpawnDetached(ctx, name, c.Config, spawnTimeout, true)
	paths, perr := daemon.PathsFor(name)
	if perr != nil {
		return fmt.Errorf("stack: %w", perr)
	}
	switch {
	case errors.Is(err, daemon.ErrAlreadyRunning):
		// already running; attach silently.
	case errors.Is(err, daemon.ErrNameConflict):
		return fmt.Errorf("stack: %w (pass a different name to run a second instance)", err)
	case err != nil:
		return fmt.Errorf("stack: %w", err)
	default:
		fmt.Printf("%s started\n", name)
	}

	if c.Detach {
		return nil
	}
	return tui.Run(name, paths.Socket, paths.DaemonLog, true)
}
