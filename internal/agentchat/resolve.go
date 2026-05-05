package agentchat

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/iorubs/smithy-cli/internal/config"
	"github.com/iorubs/smithy-cli/internal/runtime"
	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
)

// AgentTarget is the resolved coordinates a chat client needs to
// reach a daemon-supervised agent.
type AgentTarget struct {
	Name      string
	StackName string
	Transport string
	BaseURL   string
}

// Resolve finds an agent in the running stack, validates it speaks a
// supported chat transport, and returns its base URL.
func Resolve(ctx context.Context, stackName, agentName string) (AgentTarget, error) {
	paths, err := daemon.PathsFor(stackName)
	if err != nil {
		return AgentTarget{}, err
	}
	if !daemon.IsStackRunning(ctx, paths, 2*time.Second) {
		return AgentTarget{}, fmt.Errorf("stack %q is not running (start it with `smithy stack up`)", stackName)
	}

	meta, ok := daemon.ReadMeta(paths)
	if !ok {
		return AgentTarget{}, fmt.Errorf("stack %q: cannot read metadata", stackName)
	}

	data, err := os.ReadFile(meta.ConfigPath)
	if err != nil {
		return AgentTarget{}, fmt.Errorf("read stack file %s: %w", meta.ConfigPath, err)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		return AgentTarget{}, fmt.Errorf("parse stack file: %w", err)
	}
	plan, err := runtime.Translate(cfg, meta.ConfigPath)
	if err != nil {
		return AgentTarget{}, fmt.Errorf("translate stack: %w", err)
	}

	for _, spec := range plan.Agents {
		if spec.Name != agentName {
			continue
		}
		transport := spec.Transport
		if transport == "" {
			transport = "a2a"
		}
		if transport != "a2a" && transport != "mcp-http" {
			return AgentTarget{}, fmt.Errorf(
				"agent %q uses transport %q; chat currently supports a2a and mcp-http only",
				agentName, transport)
		}
		addr := spec.Addr
		if addr == "" {
			addr = ":8081"
		}
		return AgentTarget{
			Name:      agentName,
			StackName: stackName,
			Transport: transport,
			BaseURL:   addrToURL(addr),
		}, nil
	}
	return AgentTarget{}, fmt.Errorf("agent %q not declared in stack %q", agentName, stackName)
}

// addrToURL turns a listen address (":8081", "0.0.0.0:8081",
// "127.0.0.1:8081") into a reachable http URL with a trailing slash,
// matching the AgentCard URL the server publishes.
func addrToURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%s/", host, port)
}
