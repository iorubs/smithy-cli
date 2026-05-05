package runtime

import (
	"context"
	"io"

	"github.com/iorubs/agentsmithy/pkg/api"
)

// defaultAgentAddr mirrors the agentsmithy CLI's `--addr` default.
// Applied here when an agent entry leaves `addr` unset on a transport
// that requires one (a2a, mcp-http).
const defaultAgentAddr = ":8081"

// RunAgent is the default AgentRunner. It loads the referenced
// agentsmithy config and delegates to api.Serve.
func RunAgent(ctx context.Context, spec AgentSpec, _, _ io.Writer) error {
	ctx = WithServiceKind(ctx, spec.Name, "agent")
	cfg, root, err := api.LoadConfig(spec.ConfigPath)
	if err != nil {
		return err
	}
	transport := spec.Transport
	if transport == "" {
		transport = "a2a"
	}
	addr := spec.Addr
	if addr == "" && (transport == "a2a" || transport == "mcp-http") {
		addr = defaultAgentAddr
	}
	return api.Serve(ctx, cfg, api.ServeOptions{
		Root:      root,
		Transport: transport,
		Addr:      addr,
	})
}
