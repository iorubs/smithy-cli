package runtime

import (
	"context"
	"io"

	"github.com/iorubs/mcpsmithy/pkg/api"
)

// defaultHTTPAddr mirrors the mcpsmithy CLI's `--addr` default. The
// pkg/api layer doesn't apply it, so we apply it here when an MCP
// entry leaves `addr` unset on an http transport.
const defaultHTTPAddr = ":8080"

// RunMCP is the default Runner. It loads the referenced mcpsmithy
// config and delegates to api.Serve.
func RunMCP(ctx context.Context, spec MCPSpec, _, _ io.Writer) error {
	ctx = WithServiceKind(ctx, spec.Name, "mcp")
	cfg, root, err := api.LoadConfig(spec.ConfigPath)
	if err != nil {
		return err
	}
	addr := spec.Addr
	if addr == "" && spec.Transport == "http" {
		addr = defaultHTTPAddr
	}
	return api.Serve(ctx, cfg, api.ServeOptions{
		Root:      root,
		Transport: spec.Transport,
		Addr:      addr,
	})
}
