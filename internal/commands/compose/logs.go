package compose

import "context"

// LogsCmd streams logs from one or more servers.
type LogsCmd struct {
}

// Run executes the logs command.
func (c *LogsCmd) Run(ctx context.Context) error {
	return nil
}
