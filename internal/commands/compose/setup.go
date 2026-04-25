package compose

import "context"

// SetupCmd initialises a new compose project in the current directory.
type SetupCmd struct {
}

// Run executes the setup command.
func (c *SetupCmd) Run(ctx context.Context) error {
	return nil
}
