// Package agentchat: MCP-HTTP transport client.
//
// Each chat turn is sent as a single tools/call against the agent's
// exposed pipeline tool. The MCP server keeps no chat history of its
// own across turns (each invocation is a fresh agent session under
// the hood), so FetchHistory always returns nil. This is fine for a
// first cut — the round-trip still works the same way as a2a from
// the user's POV.
package agentchat

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mcpHTTPClient holds an open MCP session against an agentsmithy
// server running with --transport mcp-http. The session is opened
// once and reused for every turn.
type mcpHTTPClient struct {
	target   AgentTarget
	session  *mcp.ClientSession
	toolName string

	mu sync.Mutex
}

// newMCPHTTPClient connects to the agent's MCP HTTP endpoint and
// resolves the (single) tool the agentsmithy server exposes for
// pipeline invocations. agentsmithy registers exactly one tool per
// pipeline whose name is a sanitized form of the pipeline name; we
// just take the first tool the server lists.
func newMCPHTTPClient(ctx context.Context, target AgentTarget) (*mcpHTTPClient, error) {
	cli := mcp.NewClient(&mcp.Implementation{
		Name:    "smithy-cli",
		Version: "1.0.0",
	}, nil)

	endpoint := strings.TrimSuffix(target.BaseURL, "/")
	transport := &mcp.StreamableClientTransport{
		Endpoint:   endpoint,
		HTTPClient: nil,
	}
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	session, err := cli.Connect(connectCtx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp connect %s: %w", endpoint, err)
	}

	tools, err := session.ListTools(connectCtx, nil)
	if err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("mcp list tools: %w", err)
	}
	if len(tools.Tools) == 0 {
		_ = session.Close()
		return nil, fmt.Errorf("mcp server at %s exposes no tools", endpoint)
	}
	return &mcpHTTPClient{
		target:   target,
		session:  session,
		toolName: tools.Tools[0].Name,
	}, nil
}

// SendText invokes the pipeline tool with the user prompt and returns
// the agent's reply text.
func (c *mcpHTTPClient) SendText(ctx context.Context, text string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	res, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      c.toolName,
		Arguments: map[string]any{"prompt": text},
	})
	if err != nil {
		return "", fmt.Errorf("mcp call %s: %w", c.toolName, err)
	}
	if res.IsError {
		return "", fmt.Errorf("mcp call %s: %s", c.toolName, contentText(res.Content))
	}
	return contentText(res.Content), nil
}

// FetchHistory returns nil; mcp-http has no per-conversation history
// to replay. Existing chat sessions show only turns produced in the
// current process.
func (c *mcpHTTPClient) FetchHistory(ctx context.Context) ([]Turn, error) {
	return nil, nil
}

// Close terminates the MCP session.
func (c *mcpHTTPClient) Close() error {
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}

func contentText(content []mcp.Content) string {
	var b strings.Builder
	for _, c := range content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return strings.TrimSpace(b.String())
}
