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
	// EnvFile is an optional list of dotenv files (KEY=VALUE) loaded
	// into the daemon process before any service starts. Paths are
	// resolved relative to the stack file's directory. Compose-style
	// semantics: the shell environment wins over file values; among
	// declared files, the first listed wins. A `.env` next to the
	// stack file is loaded automatically when this list is empty.
	EnvFile []string `yaml:"env_file,omitempty"`
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
	// Listen address (host:port) for http transports. Mirrors
	// mcpsmithy's --addr flag. Ignored for stdio.
	Addr string `yaml:"addr,omitempty"`
	// Restart on non-zero exit. Defaults to true when unset; set false
	// to opt out per-entry (one-shot or hand-managed servers).
	AutoRestart *bool `yaml:"autorestart,omitempty"`
}

// Agent is a single agent entry. The referenced config is loaded by
// agentsmithy and served in-process by the supervisor.
type Agent struct {
	// Path to the agentsmithy config file (.agentsmithy.yaml).
	// Resolved relative to the stack file's directory.
	Config string `yaml:"config,omitempty" smithy:"required"`
	// Wire transport exposed by the agent. Defaults to a2a when unset.
	Transport AgentTransport `yaml:"transport,omitempty"`
	// Listen address (host:port) for http-like transports. Mirrors
	// agentsmithy's --addr flag. Ignored for stdio transports.
	Addr string `yaml:"addr,omitempty"`
	// Restart on non-zero exit. Defaults to true when unset; set false
	// to opt out per-entry (one-shot or hand-managed agents).
	AutoRestart *bool `yaml:"autorestart,omitempty"`
}

// Transport selects the wire protocol an MCP server exposes.
type Transport string

const (
	// TransportStdio runs the server over stdio.
	TransportStdio Transport = "stdio"
	// TransportHTTP serves the server over streamable HTTP.
	TransportHTTP Transport = "http"
)

// Values lists the allowed transport strings.
func (Transport) Values() []string {
	return []string{string(TransportStdio), string(TransportHTTP)}
}

// AgentTransport selects the wire protocol an agent exposes.
type AgentTransport string

const (
	// AgentTransportA2A serves the agent as A2A JSON-RPC over HTTP.
	AgentTransportA2A AgentTransport = "a2a"
	// AgentTransportMCPHTTP serves the agent as an MCP server over
	// streamable HTTP.
	AgentTransportMCPHTTP AgentTransport = "mcp-http"
	// AgentTransportMCPStdio serves the agent as an MCP server over
	// stdio.
	AgentTransportMCPStdio AgentTransport = "mcp-stdio"
	// AgentTransportStdio runs the agent's line REPL over stdio.
	AgentTransportStdio AgentTransport = "stdio"
)

// Values lists the allowed agent transport strings.
func (AgentTransport) Values() []string {
	return []string{
		string(AgentTransportA2A),
		string(AgentTransportMCPHTTP),
		string(AgentTransportMCPStdio),
		string(AgentTransportStdio),
	}
}
