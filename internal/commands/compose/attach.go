package compose

// AttachCmd attaches to a running server's stdio for interactive use.
type AttachCmd struct {
}

// Run executes the attach command.
func (c *AttachCmd) Run() error {
	return nil
}
