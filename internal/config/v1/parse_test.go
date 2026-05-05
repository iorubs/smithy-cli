package v1

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	trueVal, falseVal := true, false
	tests := []struct {
		name    string
		yaml    string
		wantErr string
		check   func(t *testing.T, c *Config)
	}{
		{
			name: "version only",
			yaml: `version: "1"`,
		},
		{
			name: "mcp only",
			yaml: `
version: "1"
mcps:
  fetch:
    config: ./fetch.mcpsmithy.yaml
    transport: http
    addr: 127.0.0.1:8080
`,
			check: func(t *testing.T, c *Config) {
				m, ok := c.MCPs["fetch"]
				if !ok {
					t.Fatalf("missing mcps.fetch")
				}
				if m.Config != "./fetch.mcpsmithy.yaml" {
					t.Errorf("config = %q", m.Config)
				}
				if m.Transport != TransportHTTP {
					t.Errorf("transport = %q", m.Transport)
				}
				if m.Addr != "127.0.0.1:8080" {
					t.Errorf("addr = %q", m.Addr)
				}
				if m.AutoRestart != nil {
					t.Errorf("autorestart = %v, want nil", *m.AutoRestart)
				}
			},
		},
		{
			name: "autorestart explicit false round-trips",
			yaml: `
version: "1"
mcps:
  one:
    config: ./a.yaml
    autorestart: false
  two:
    config: ./b.yaml
    autorestart: true
`,
			check: func(t *testing.T, c *Config) {
				if got := c.MCPs["one"].AutoRestart; got == nil || *got != falseVal {
					t.Errorf("one.autorestart = %v, want ptr(false)", got)
				}
				if got := c.MCPs["two"].AutoRestart; got == nil || *got != trueVal {
					t.Errorf("two.autorestart = %v, want ptr(true)", got)
				}
			},
		},
		{
			name: "missing required mcp config",
			yaml: `
version: "1"
mcps:
  bad:
    transport: http
`,
			wantErr: "is required",
		},
		{
			name: "agent only",
			yaml: `
version: "1"
agents:
  primary:
    config: ./.agentsmithy.yaml
    transport: a2a
    addr: 127.0.0.1:8081
`,
			check: func(t *testing.T, c *Config) {
				a, ok := c.Agents["primary"]
				if !ok {
					t.Fatalf("missing agents.primary")
				}
				if a.Config != "./.agentsmithy.yaml" {
					t.Errorf("config = %q", a.Config)
				}
				if a.Transport != AgentTransportA2A {
					t.Errorf("transport = %q", a.Transport)
				}
				if a.Addr != "127.0.0.1:8081" {
					t.Errorf("addr = %q", a.Addr)
				}
			},
		},
		{
			name: "missing required agent config",
			yaml: `
version: "1"
agents:
  bad:
    transport: a2a
`,
			wantErr: "is required",
		},
		{
			name: "invalid agent transport",
			yaml: `
version: "1"
agents:
  bad:
    config: ./a.yaml
    transport: grpc
`,
			wantErr: "must be one of",
		},
		{
			name: "invalid transport",
			yaml: `
version: "1"
mcps:
  bad:
    config: ./a.yaml
    transport: grpc
`,
			wantErr: "must be one of",
		},
		{
			name: "unknown field rejected",
			yaml: `
version: "1"
mcps:
  bad:
    config: ./a.yaml
    bogus: nope
`,
			wantErr: "bogus",
		},
		{
			name:    "missing version",
			yaml:    `mcps: {}`,
			wantErr: "is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Schema{}.Parse([]byte(tt.yaml))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}
