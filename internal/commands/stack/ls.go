package stack

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
)

// LsCmd lists every stack with a daemon directory under the user's
// stacks root, marking each as running or stale.
type LsCmd struct{}

// Run executes the ls command.
func (c *LsCmd) Run(ctx context.Context) error {
	names, err := daemon.ListNames()
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATE\tPID\tCONFIG")
	for _, n := range names {
		paths, err := daemon.PathsFor(n)
		if err != nil {
			continue
		}
		state := "stale"
		if daemon.IsStackRunning(ctx, paths, 500*time.Millisecond) {
			state = "running"
		}
		pid := ""
		config := ""
		if meta, ok := daemon.ReadMeta(paths); ok {
			pid = fmt.Sprintf("%d", meta.PID)
			config = meta.ConfigPath
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", n, state, pid, config)
	}
	return tw.Flush()
}
