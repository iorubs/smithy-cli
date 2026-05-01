package stack

import (
	"context"
	"log/slog"

	"github.com/iorubs/smithy-cli/internal/setup"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SetupCmd starts a stdio MCP server that helps an LLM author a
// smithy-stack.yaml. It does not require an existing stack file.
type SetupCmd struct{}

// Run starts the setup server on stdio.
func (c *SetupCmd) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "setup server running on stdio; connect your agent to write smithy-stack.yaml")
	slog.InfoContext(ctx, "when done: smithy stack validate; then: smithy stack up")
	srv := setup.BuildServer()
	return srv.Run(ctx, &mcp.StdioTransport{})
}
