package stack

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

// PsCmd lists services in a running stack by querying its /status
// socket. The stack can be addressed by positional name or by
// -c <file>. Exits non-zero if the stack is not running.
type PsCmd struct {
	ConfigFlag
	Name string `arg:"" optional:"" help:"Stack name."`
}

// Run executes the ps command.
func (c *PsCmd) Run(ctx context.Context) error {
	name, err := ResolveStackName(c.Name, c.Config)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}
	paths, err := daemon.PathsFor(name)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}
	if !daemon.IsStackRunning(ctx, paths, 2*time.Second) {
		return fmt.Errorf("stack: stack %q is not running", name)
	}
	resp, err := ipc.NewClient(paths.Socket).Status(ctx)
	if err != nil {
		return fmt.Errorf("stack: query daemon: %w", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tKIND\tSTATE")
	for _, s := range resp.Services {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", s.Name, s.Kind, s.State)
	}
	return tw.Flush()
}
