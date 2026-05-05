package stack

import (
	"context"
	"fmt"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

// downStopTimeout bounds how long down waits for the daemon to ack
// shutdown and tear its socket down.
const downStopTimeout = 15 * time.Second

// DownCmd stops a backgrounded stack started with `stack up`. The
// stack can be addressed by positional name or by -c <file> (which
// derives the same name as `up -c <file>` would have used).
type DownCmd struct {
	ConfigFlag
	Name string `arg:"" optional:"" help:"Stack name."`
}

// Run executes the down command.
func (c *DownCmd) Run(ctx context.Context) error {
	name, err := ResolveStackName(c.Name, c.Config)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}
	paths, err := daemon.PathsFor(name)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}

	if !daemon.IsStackRunning(ctx, paths, 500*time.Millisecond) {
		daemon.CleanupArtifacts(paths)
		return nil
	}

	client := ipc.NewClient(paths.Socket)
	shutCtx, cancel := context.WithTimeout(ctx, downStopTimeout)
	defer cancel()

	if err := client.Shutdown(shutCtx); err != nil {
		fmt.Printf("%s: socket shutdown failed, sending signal\n", name)
		if sigErr := daemon.SignalFromPID(paths); sigErr != nil {
			return fmt.Errorf("stack: stop daemon: %w", sigErr)
		}
	}

	if err := daemon.WaitForExit(shutCtx, paths); err != nil {
		return fmt.Errorf("stack: %w", err)
	}

	daemon.CleanupArtifacts(paths)
	fmt.Printf("%s: stopped\n", name)
	return nil
}
