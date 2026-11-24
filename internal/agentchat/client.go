package agentchat

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
)

// a2aV1Path is where agentsmithy serves the A2A v1.0 JSON-RPC binding.
const a2aV1Path = "/a2a/v1"

// Turn is one entry in the chat transcript.
type Turn struct {
	From string // "you" or "agent"
	Text string
}

// a2aClient wraps a2aclient.Client and persists the a2a contextID
// inside the stack's daemon metadata across CLI invocations.
// Transcript history is fetched from the agent (via tasks/list) on
// demand; the agent's session service is the source of truth.
type a2aClient struct {
	target    AgentTarget
	a2a       *a2aclient.Client
	contextID string
	paths     daemon.Paths
}

// newA2AClient builds an a2a client for target and loads any persisted
// contextID for the (stack, agent) pair from stack metadata. When
// reset is true the entry is removed so the next send starts a fresh
// conversation server-side.
func newA2AClient(ctx context.Context, target AgentTarget, reset bool) (*a2aClient, error) {
	paths, err := daemon.PathsFor(target.StackName)
	if err != nil {
		return nil, err
	}
	if reset {
		if err := daemon.SetChatContextID(paths, target.Name, ""); err != nil {
			slog.WarnContext(ctx, "agentchat: clear chat context", "agent", target.Name, "error", err)
		}
	}
	contextID := ""
	if meta, ok := daemon.ReadMeta(paths); ok {
		contextID = meta.Chats[target.Name]
	}

	// agentsmithy serves A2A v1.0 JSON-RPC at /a2a/v1 (v0.3 stays at /
	// for older clients). Agents run in-process from the same
	// agentsmithy version, so the v1.0 endpoint is always available.
	rpcURL := strings.TrimSuffix(target.BaseURL, "/") + a2aV1Path
	card := &a2a.AgentCard{
		Name:    target.Name,
		Version: "1.0.0",
		SupportedInterfaces: []*a2a.AgentInterface{
			{URL: rpcURL, ProtocolBinding: a2a.TransportProtocolJSONRPC, ProtocolVersion: a2a.Version},
		},
	}
	factory := a2aclient.NewFactory(
		a2aclient.WithJSONRPCTransport(&http.Client{Timeout: 5 * time.Minute}),
	)
	a2aCli, err := factory.CreateFromCard(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("a2a client: %w", err)
	}
	return &a2aClient{
		target:    target,
		a2a:       a2aCli,
		contextID: contextID,
		paths:     paths,
	}, nil
}

// Close releases any resources held by the underlying a2a client.
func (c *a2aClient) Close() error { return c.a2a.Destroy() }

// ContextID returns the current conversation context, or "" if no
// reply has been received yet.
func (c *a2aClient) ContextID() string { return c.contextID }

// FetchHistory pulls the conversation transcript from the agent for
// the persisted contextID via agentsmithy's GET /sessions/{id}/messages
// endpoint. Returns an empty slice when there is no contextID yet or
// the agent has no events for it.
//
// Note: a2a's standard tasks/list RPC is gated behind an
// authenticator hook in a2a-go's in-memory store; agentsmithy
// exposes the session events directly over a small REST shim so the
// CLI can recover history without an auth setup.
func (c *a2aClient) FetchHistory(ctx context.Context) ([]Turn, error) {
	if c.contextID == "" {
		return nil, nil
	}
	url := strings.TrimSuffix(c.target.BaseURL, "/") + "/sessions/" + c.contextID + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch history: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch history: %s", resp.Status)
	}
	var body struct {
		Messages []struct {
			Role string `json:"role"`
			Text string `json:"text"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode history: %w", err)
	}
	out := make([]Turn, 0, len(body.Messages))
	for _, m := range body.Messages {
		from := "agent"
		if m.Role == "user" {
			from = "you"
		}
		out = append(out, Turn{From: from, Text: m.Text})
	}
	return out, nil
}

// SendText sends one user turn and returns the agent's text reply.
// On the first successful round-trip, the server-issued contextID is
// captured and persisted into the stack metadata file so subsequent
// calls (including across process restarts) reattach to the same
// conversation. The metadata file is owned by the daemon and removed
// on stack shutdown, so contextIDs auto-clear when the in-memory
// agent sessions they reference are gone too.
func (c *a2aClient) SendText(ctx context.Context, text string) (string, error) {
	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(text))
	msg.ContextID = c.contextID
	resp, err := c.a2a.SendMessage(ctx, &a2a.SendMessageRequest{Message: msg})
	if err != nil {
		return "", fmt.Errorf("a2a send: %w", err)
	}
	if cid := contextIDOf(resp); cid != "" && cid != c.contextID {
		c.contextID = cid
		if err := daemon.SetChatContextID(c.paths, c.target.Name, cid); err != nil {
			slog.WarnContext(ctx, "agentchat: persist chat context", "agent", c.target.Name, "error", err)
		}
	}
	return extractText(resp), nil
}

func contextIDOf(resp a2a.SendMessageResult) string {
	switch v := resp.(type) {
	case *a2a.Message:
		return v.ContextID
	case *a2a.Task:
		return v.ContextID
	}
	return ""
}

func extractText(resp a2a.SendMessageResult) string {
	switch v := resp.(type) {
	case *a2a.Message:
		return textOfParts(v.Parts)
	case *a2a.Task:
		for _, art := range v.Artifacts {
			if t := textOfParts(art.Parts); t != "" {
				return t
			}
		}
		if v.Status.Message != nil {
			return textOfParts(v.Status.Message.Parts)
		}
	}
	return ""
}

func textOfParts(parts a2a.ContentParts) string {
	var b strings.Builder
	for _, p := range parts {
		if p != nil {
			b.WriteString(p.Text())
		}
	}
	return b.String()
}
