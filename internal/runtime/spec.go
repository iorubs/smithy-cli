// Package runtime supervises smithy-stack stacks.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	v1 "github.com/iorubs/smithy-cli/internal/config/v1"
)

// Plan is the resolved set of services Launch will run.
type Plan struct {
	MCPs   []MCPSpec
	Agents []AgentSpec
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
	// Addr is the listen address ("host:port") for http transports;
	// empty for stdio or when the stack entry didn't set it.
	Addr string
	// AutoRestart is the resolved restart policy (nil → true).
	AutoRestart bool
}

// AgentSpec is a single agent resolved from the stack schema.
type AgentSpec struct {
	// Name is the stack-file key for this entry.
	Name string
	// ConfigPath is the absolute, existence-checked path to the
	// referenced agentsmithy config file.
	ConfigPath string
	// Transport is the wire protocol the agent should expose. Empty
	// means defer to agentsmithy's default (a2a).
	Transport string
	// Addr is the listen address ("host:port") for http-like
	// transports; empty for stdio variants or when unset.
	Addr string
	// AutoRestart is the resolved restart policy (nil → true).
	AutoRestart bool
}

// Runner runs one MCP attempt and returns when ctx is cancelled or
// the server exits. Launch fans out to a Runner per spec; tests inject
// fakes here to exercise fan-out behaviour without booting real MCPs.
type Runner func(ctx context.Context, spec MCPSpec, stdout, stderr io.Writer) error

// AgentRunner runs one agent attempt; mirrors Runner for the agent
// side of the plan.
type AgentRunner func(ctx context.Context, spec AgentSpec, stdout, stderr io.Writer) error

// Translate resolves a parsed stack Config into a Plan suitable for
// Launch. stackPath is the path the Config was parsed from; relative
// MCP and agent config paths are resolved against its directory.
// Translate also existence-checks each referenced config file so
// typos surface in `stack validate` rather than at launch time.
func Translate(cfg *v1.Config, stackPath string) (Plan, error) {
	if cfg == nil {
		return Plan{}, fmt.Errorf("translate: config is nil")
	}
	stackDir, err := filepath.Abs(filepath.Dir(stackPath))
	if err != nil {
		return Plan{}, fmt.Errorf("translate: resolving stack dir: %w", err)
	}

	mcpNames := make([]string, 0, len(cfg.MCPs))
	for name := range cfg.MCPs {
		mcpNames = append(mcpNames, name)
	}
	sort.Strings(mcpNames)

	mcpSpecs := make([]MCPSpec, 0, len(mcpNames))
	for _, name := range mcpNames {
		mcp := cfg.MCPs[name]

		path, err := resolveConfigPath(stackDir, mcp.Config)
		if err != nil {
			return Plan{}, fmt.Errorf("mcp %q: %w", name, err)
		}

		autoRestart := true
		if mcp.AutoRestart != nil {
			autoRestart = *mcp.AutoRestart
		}

		mcpSpecs = append(mcpSpecs, MCPSpec{
			Name:        name,
			ConfigPath:  path,
			Transport:   string(mcp.Transport),
			Addr:        mcp.Addr,
			AutoRestart: autoRestart,
		})
	}

	agentNames := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		agentNames = append(agentNames, name)
	}
	sort.Strings(agentNames)

	agentSpecs := make([]AgentSpec, 0, len(agentNames))
	for _, name := range agentNames {
		agent := cfg.Agents[name]

		path, err := resolveConfigPath(stackDir, agent.Config)
		if err != nil {
			return Plan{}, fmt.Errorf("agent %q: %w", name, err)
		}

		autoRestart := true
		if agent.AutoRestart != nil {
			autoRestart = *agent.AutoRestart
		}

		agentSpecs = append(agentSpecs, AgentSpec{
			Name:        name,
			ConfigPath:  path,
			Transport:   string(agent.Transport),
			Addr:        agent.Addr,
			AutoRestart: autoRestart,
		})
	}

	return Plan{MCPs: mcpSpecs, Agents: agentSpecs}, nil
}

// resolveConfigPath returns the absolute, existence-checked path for
// a stack-file config reference. Relative paths resolve against
// stackDir.
func resolveConfigPath(stackDir, configPath string) (string, error) {
	path := configPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(stackDir, path)
	}
	path = filepath.Clean(path)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%s not found", configPath)
		}
		return "", fmt.Errorf("%s: %w", configPath, err)
	}
	return path, nil
}
