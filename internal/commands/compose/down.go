package compose

// DownCmd stops all servers in the compose file and the supervisor daemon.
type DownCmd struct {
}

// Run executes the down command.
func (c *DownCmd) Run() error {
	return nil
}
