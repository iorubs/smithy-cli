package compose

// UpCmd starts all servers declared in the compose file.
type UpCmd struct {
}

// Run executes the up command.
func (c *UpCmd) Run() error {
	return nil
}
