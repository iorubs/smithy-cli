package setup

import (
	"context"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	eng := New()
	// Both tools must be reachable via Execute.
	if _, err := eng.Execute(context.Background(), toolConfigGuide, nil); err != nil {
		t.Errorf("config_guide: %v", err)
	}
	if _, err := eng.Execute(context.Background(), toolConfigSection, map[string]any{"section": "mcps"}); err != nil {
		t.Errorf("config_section: %v", err)
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name         string
		tool         string
		params       map[string]any
		wantErr      bool
		wantContains []string
	}{
		// config_guide
		{"guide returns overview", toolConfigGuide, nil, false, []string{"smithy-stack", "config_section"}},

		// config_section: valid
		{"section mcps", toolConfigSection, map[string]any{"section": "mcps"}, false, []string{"mcps Section", "Field Reference", "## MCP", "## Transport"}},
		{"section agents", toolConfigSection, map[string]any{"section": "agents"}, false, []string{"agents Section", "Field Reference", "## Agent"}},
		{"case insensitive upper", toolConfigSection, map[string]any{"section": "MCPS"}, false, []string{"mcps Section"}},
		{"case insensitive mixed", toolConfigSection, map[string]any{"section": "Agents"}, false, []string{"agents Section"}},

		// config_section: errors
		{"unknown section", toolConfigSection, map[string]any{"section": "banana"}, true, []string{"banana"}},
		{"empty section key", toolConfigSection, map[string]any{}, true, []string{"section"}},
		{"nil params", toolConfigSection, nil, true, []string{"section"}},

		// unknown tool
		{"unknown tool", "nonexistent", nil, true, []string{"nonexistent"}},
	}
	eng := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := eng.Execute(context.Background(), tt.tool, tt.params)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				for _, s := range tt.wantContains {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("error should mention %q, got: %v", s, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == "" {
				t.Fatal("expected non-empty output")
			}
			for _, s := range tt.wantContains {
				if !strings.Contains(out, s) {
					t.Errorf("output missing %q", s)
				}
			}
		})
	}
}

func TestSectionsHaveContent(t *testing.T) {
	sections := append([]string{"guide"}, validSections...)
	for _, name := range sections {
		t.Run(name, func(t *testing.T) {
			content, err := readSection(name)
			if err != nil {
				t.Fatalf("failed to read section %q: %v", name, err)
			}
			if len(content) < 100 {
				t.Errorf("section %q seems too short (%d bytes)", name, len(content))
			}
		})
	}
}
