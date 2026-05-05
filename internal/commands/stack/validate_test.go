package stack

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCmdRun(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid stack file",
			yaml: `version: "1"
mcps:
  fetch:
    config: ./fetch.mcpsmithy.yaml
    transport: http
    addr: 127.0.0.1:8080
`,
		},
		{
			name: "missing required version",
			yaml: `mcps:
  fetch:
    config: ./fetch.mcpsmithy.yaml
`,
			wantErr: "unsupported config version",
		},
		{
			name: "missing required mcp config field",
			yaml: `version: "1"
mcps:
  fetch:
    transport: http
`,
			wantErr: "config",
		},
		{
			name: "bad transport enum value",
			yaml: `version: "1"
mcps:
  fetch:
    config: ./fetch.mcpsmithy.yaml
    transport: tcp
`,
			wantErr: "transport",
		},
		{
			name: "unknown field rejected",
			yaml: `version: "1"
mcps:
  fetch:
    config: ./fetch.mcpsmithy.yaml
    bogus: nope
`,
			wantErr: "bogus",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "smithy-stack.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o600); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			if err := os.WriteFile(filepath.Join(dir, "fetch.mcpsmithy.yaml"), []byte("project:\n  name: x\n"), 0o600); err != nil {
				t.Fatalf("write referenced config: %v", err)
			}
			cmd := &ValidateCmd{ConfigFlag: ConfigFlag{Config: path}}
			err := cmd.Run(context.Background())
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCmdMissingFile(t *testing.T) {
	cmd := &ValidateCmd{ConfigFlag: ConfigFlag{Config: filepath.Join(t.TempDir(), "does-not-exist.yaml")}}
	err := cmd.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
