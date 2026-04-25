package compose

// RestartCmd restarts one or more managed servers.
type RestartCmd struct {
}

// Run executes the restart command.
func (c *RestartCmd) Run() error {
	return nil
}
