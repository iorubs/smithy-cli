package agentchat

import (
	"context"

	v1 "github.com/iorubs/smithy-cli/internal/config/v1"
)

// Client is the chat surface the TUI consumes. Two transports are
// supported today: A2A (default; persists conversation context across
// CLI runs and can replay prior history) and MCP-HTTP (each turn is a
// fresh tool call; no history). Other transports (mcp-stdio, stdio)
// require daemon-side relays and are not yet wired.
type Client interface {
	SendText(ctx context.Context, text string) (string, error)
	FetchHistory(ctx context.Context) ([]Turn, error)
	Close() error
}

// NewClient picks the right transport implementation for target.
func NewClient(ctx context.Context, target AgentTarget, reset bool) (Client, error) {
	switch target.Transport {
	case string(v1.AgentTransportMCPHTTP):
		return newMCPHTTPClient(ctx, target)
	default:
		return newA2AClient(ctx, target, reset)
	}
}
