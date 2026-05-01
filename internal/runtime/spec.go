// Package runtime supervises smithy-stack stacks.
package runtime

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	v1 "github.com/iorubs/smithy-cli/internal/config/v1"
)

// Plan is the resolved set of services Launch will run.
type Plan struct {
	MCPs []MCPSpec
}

// MCPSpec is a single MCP server resolved from the stack schema.
type MCPSpec struct {
	// Name is the stack-file key for this entry.
	Name string
	// ConfigPath is the absolute, existence-checked path to the referenced mcpsmithy config file.
	ConfigPath string
	// Transport is the wire protocol the server should expose. Empty
	// means defer to whatever the referenced mcpsmithy config selects.
	Transport string
	// Addr is the listen address ("host:port") for http/sse transports;
	// empty for stdio or when the stack entry didn't set it.
	Addr string
	// AutoRestart is the resolved restart policy (nil → true).
	AutoRestart bool
}

// Runner runs one MCP attempt and returns when ctx is cancelled or
// the server exits. Launch fans out to a Runner per spec; tests inject
// fakes here to exercise fan-out behaviour without booting real MCPs.
type Runner func(ctx context.Context, spec MCPSpec, stdout, stderr io.Writer) error

// Translate resolves a parsed stack Config into a Plan suitable for
// Launch. stackPath is the path the Config was parsed from; relative
// MCP config paths are resolved against its directory. Translate also
// existence-checks each referenced config file so typos surface in
// `stack validate` rather than at launch time.
func Translate(cfg *v1.Config, stackPath string) (Plan, error) {
	if cfg == nil {
		return Plan{}, fmt.Errorf("translate: config is nil")
	}
	stackDir, err := filepath.Abs(filepath.Dir(stackPath))
	if err != nil {
		return Plan{}, fmt.Errorf("translate: resolving stack dir: %w", err)
	}

	names := make([]string, 0, len(cfg.MCPs))
	for name := range cfg.MCPs {
		names = append(names, name)
	}
	sort.Strings(names)

	specs := make([]MCPSpec, 0, len(names))
	for _, name := range names {
		mcp := cfg.MCPs[name]

		path := mcp.Config
		if !filepath.IsAbs(path) {
			path = filepath.Join(stackDir, path)
		}
		path = filepath.Clean(path)
		if _, err := os.Stat(path); err != nil {
			return Plan{}, fmt.Errorf("mcp %q: config: %w", name, err)
		}

		autoRestart := true
		if mcp.AutoRestart != nil {
			autoRestart = *mcp.AutoRestart
		}

		specs = append(specs, MCPSpec{
			Name:        name,
			ConfigPath:  path,
			Transport:   string(mcp.Transport),
			Addr:        mcp.Addr,
			AutoRestart: autoRestart,
		})
	}

	return Plan{MCPs: specs}, nil
}
