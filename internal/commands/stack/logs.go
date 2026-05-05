package stack

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
	"github.com/iorubs/smithy-cli/internal/tui"
)

// logsFollowPollInterval is how often -f re-reads after EOF.
const logsFollowPollInterval = 200 * time.Millisecond

// LogsCmd streams the daemon log for a stack. The stack can be
// addressed by positional name or by -c <file>. By default it opens
// an interactive TUI. Pass --json to stream raw log output instead;
// combined with -f that follows the file like tail -f.
type LogsCmd struct {
	ConfigFlag
	Name   string `arg:"" optional:"" help:"Stack name."`
	Follow bool   `help:"Follow log output in --json mode." short:"f"`
	JSON   bool   `help:"Output raw log instead of the interactive TUI." name:"json"`
}

// Run executes the logs command.
func (c *LogsCmd) Run(ctx context.Context) error {
	name, err := ResolveStackName(c.Name, c.Config)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}
	paths, err := daemon.PathsFor(name)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}
	if _, statErr := os.Stat(paths.DaemonLog); errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("stack: no log for stack %q", name)
	}

	if !c.JSON {
		return tui.Run(name, paths.Socket, paths.DaemonLog, false)
	}

	f, err := os.Open(paths.DaemonLog)
	if err != nil {
		return fmt.Errorf("stack: open log: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("stack: read log: %w", err)
	}
	if !c.Follow {
		return nil
	}

	for {
		_, err := io.Copy(os.Stdout, f)
		if err != nil {
			return fmt.Errorf("stack: read log: %w", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(logsFollowPollInterval):
		}
	}
}
