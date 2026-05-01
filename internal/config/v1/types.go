// Package v1 defines the v1 schema for smithy-stack.yaml.
package v1

// Version is the schema version this package handles.
const Version = "1"

// Config is the root of smithy-stack.yaml. One file declares one
// supervised stack of MCP servers and agents.
type Config struct {
	// Schema version. Must be "1".
	Version string `yaml:"version" smithy:"required"`
	// Named MCP servers managed by this stack file.
	MCPs map[string]MCP `yaml:"mcps,omitempty"`
	// Named agents managed by this stack file.
	Agents map[string]Agent `yaml:"agents,omitempty"`
}

// MCP is a single MCP server entry. The referenced config is loaded by
// mcpsmithy and served in-process by the supervisor.
type MCP struct {
	// Path to the mcpsmithy config file (.mcpsmithy.yaml). Resolved
	// relative to the stack file's directory.
	Config string `yaml:"config,omitempty" smithy:"required"`
	// Wire transport exposed by the server. Defaults to whatever the
	// referenced mcpsmithy config selects when unset.
	Transport Transport `yaml:"transport,omitempty"`
	// Listen address (host:port) for http/sse transports. Mirrors
	// mcpsmithy's --addr flag. Ignored for stdio.
	Addr string `yaml:"addr,omitempty"`
	// Restart on non-zero exit. Defaults to true when unset; set false
	// to opt out per-entry (one-shot or hand-managed servers).
	AutoRestart *bool `yaml:"autorestart,omitempty"`
}

// Agent is a single agent entry.
type Agent struct{}

// Transport selects the wire protocol an MCP server exposes.
type Transport string

const (
	// TransportStdio runs the server over stdio.
	TransportStdio Transport = "stdio"
	// TransportHTTP serves the server over streamable HTTP.
	TransportHTTP Transport = "http"
	// TransportSSE serves the server over Server-Sent Events.
	TransportSSE Transport = "sse"
)

// Values lists the allowed transport strings.
func (Transport) Values() []string {
	return []string{string(TransportStdio), string(TransportHTTP), string(TransportSSE)}
}
