package compose

import "context"

type SetupCmd struct {
}

func (c *SetupCmd) Run(ctx context.Context) error {
	return nil
}
