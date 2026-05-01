package stack

import (
	"context"
	"fmt"
	"os"

	"github.com/iorubs/smithy-cli/internal/config"
	"github.com/iorubs/smithy-cli/internal/runtime"
)

// ValidateCmd validates a stack file without starting anything.
type ValidateCmd struct {
	ConfigFlag
}

// Run parses the stack file (which runs the schema validator) and
// then translates it so cross-field errors (relative-path resolution,
// missing referenced configs) surface here instead of at launch time.
func (c *ValidateCmd) Run(_ context.Context) error {
	data, err := os.ReadFile(c.Config)
	if err != nil {
		return fmt.Errorf("stack: %w", err)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		return fmt.Errorf("stack validation: %w", err)
	}
	if _, err := runtime.Translate(cfg, c.Config); err != nil {
		return fmt.Errorf("stack validation: %w", err)
	}
	fmt.Println("config is valid")
	return nil
}
