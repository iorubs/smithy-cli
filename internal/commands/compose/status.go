package compose

// StatusCmd reports the supervisor and per-server status.
type StatusCmd struct {
}

// Run executes the status command.
func (c *StatusCmd) Run() error {
	return nil
}
