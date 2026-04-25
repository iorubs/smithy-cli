package compose

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/iorubs/smithy-cli/internal/config"
)

// ValidateCmd validates a compose file without starting anything.
type ValidateCmd struct {
	Config string `help:"Path to compose file." default:"smithy-compose.yaml" type:"path" short:"c"`
}

// Run executes the validate command.
func (c *ValidateCmd) Run(ctx context.Context) error {
	data, err := os.ReadFile(c.Config)
	if err != nil {
		return fmt.Errorf("compose: %w", err)
	}
	if _, err := config.Parse(data); err != nil {
		return fmt.Errorf("compose validation: %w", err)
	}
	slog.InfoContext(ctx, "compose file is valid", "path", c.Config)
	return nil
}
