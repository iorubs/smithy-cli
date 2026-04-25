package compose

// PsCmd lists running servers managed by the supervisor.
type PsCmd struct {
}

// Run executes the ps command.
func (c *PsCmd) Run() error {
	return nil
}
