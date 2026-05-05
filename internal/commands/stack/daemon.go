package stack

import (
	"context"
	"fmt"

	"github.com/iorubs/smithy-cli/internal/runtime"
	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
)

// DaemonCmd is the hidden subcommand the parent re-execs into when
// `stack up` runs. It is not part of the user-facing surface.
type DaemonCmd struct {
	ConfigFlag
	Name     string `help:"Stack name." required:""`
	StartAll bool   `help:"Start all services on daemon startup." name:"start-all"`
}

// Run is the daemon-mode entry point. It blocks until the parent ctx
// is cancelled (SIGTERM/SIGINT) or Launch returns.
func (c *DaemonCmd) Run(ctx context.Context) error {
	if err := daemon.Run(ctx, c.Name, c.Config, c.StartAll, runtime.LogLevelFromCtx(ctx)); err != nil {
		return fmt.Errorf("daemon: %w", err)
	}
	return nil
}
