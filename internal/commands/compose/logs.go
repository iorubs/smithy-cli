package compose

import "context"

type LogsCmd struct {
}

func (c *LogsCmd) Run(ctx context.Context) error {
	return nil
}
