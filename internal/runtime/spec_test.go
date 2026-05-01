package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/iorubs/smithy-cli/internal/config/v1"
)

func TestTranslate(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "ok.mcpsmithy.yaml")
	if err := os.WriteFile(good, []byte("project:\n  name: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stackPath := filepath.Join(dir, "smithy-stack.yaml")
	absGood, _ := filepath.Abs(good)

	cases := []struct {
		name        string
		cfg         *v1.Config
		stackPath string
		wantErr     string
		check       func(t *testing.T, plan Plan)
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "config is nil",
		},
		{
			name: "relative path resolved against stack dir",
			cfg: &v1.Config{
				Version: "1",
				MCPs:    map[string]v1.MCP{"a": {Config: "ok.mcpsmithy.yaml"}},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				if got := plan.MCPs[0].ConfigPath; got != absGood {
					t.Errorf("ConfigPath = %q, want %q", got, absGood)
				}
			},
		},
		{
			name: "absolute path passes through",
			cfg: &v1.Config{
				Version: "1",
				MCPs:    map[string]v1.MCP{"a": {Config: absGood}},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				if got := plan.MCPs[0].ConfigPath; got != absGood {
					t.Errorf("ConfigPath = %q, want %q", got, absGood)
				}
			},
		},
		{
			name: "missing config file errors",
			cfg: &v1.Config{
				Version: "1",
				MCPs:    map[string]v1.MCP{"a": {Config: "missing.yaml"}},
			},
			stackPath: stackPath,
			wantErr:     `mcp "a": config:`,
		},
		{
			name: "autorestart nil defaults to true",
			cfg: &v1.Config{
				Version: "1",
				MCPs:    map[string]v1.MCP{"a": {Config: absGood}},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				if !plan.MCPs[0].AutoRestart {
					t.Error("AutoRestart = false, want true (nil → true)")
				}
			},
		},
		{
			name: "autorestart explicit false honoured",
			cfg: &v1.Config{
				Version: "1",
				MCPs:    map[string]v1.MCP{"a": {Config: absGood, AutoRestart: new(false)}},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				if plan.MCPs[0].AutoRestart {
					t.Error("AutoRestart = true, want false")
				}
			},
		},
		{
			name: "addr passes through",
			cfg: &v1.Config{
				Version: "1",
				MCPs: map[string]v1.MCP{
					"a": {Config: absGood, Addr: "127.0.0.1:9000", Transport: v1.TransportHTTP},
				},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				if got := plan.MCPs[0].Addr; got != "127.0.0.1:9000" {
					t.Errorf("Addr = %q, want %q", got, "127.0.0.1:9000")
				}
				if got := plan.MCPs[0].Transport; got != "http" {
					t.Errorf("Transport = %q, want %q", got, "http")
				}
			},
		},
		{
			name: "empty addr stays empty",
			cfg: &v1.Config{
				Version: "1",
				MCPs:    map[string]v1.MCP{"a": {Config: absGood}},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				if got := plan.MCPs[0].Addr; got != "" {
					t.Errorf("Addr = %q, want empty", got)
				}
			},
		},
		{
			name: "specs sorted by name",
			cfg: &v1.Config{
				Version: "1",
				MCPs: map[string]v1.MCP{
					"c": {Config: absGood},
					"a": {Config: absGood},
					"b": {Config: absGood},
				},
			},
			stackPath: stackPath,
			check: func(t *testing.T, plan Plan) {
				names := []string{plan.MCPs[0].Name, plan.MCPs[1].Name, plan.MCPs[2].Name}
				want := []string{"a", "b", "c"}
				for i := range want {
					if names[i] != want[i] {
						t.Errorf("MCPs not sorted: %v, want %v", names, want)
						return
					}
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := Translate(tc.cfg, tc.stackPath)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, plan)
			}
		})
	}
}
