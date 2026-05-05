package tui

import (
	"strings"
	"testing"
)

func TestPadRight(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"toolong", 4, "tool"},
		{"", 3, "   "},
	}
	for _, tc := range tests {
		got := padRight(tc.in, tc.n)
		if got != tc.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
		}
	}
}

func TestPrettifyLogs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		svcKind     map[string]string
		wantContain []string
		wantAbsent  []string
		check       func(t *testing.T, got string)
	}{
		{
			name:  "empty input returns empty",
			input: "",
			check: func(t *testing.T, got string) {
				t.Helper()
				if got != "" {
					t.Errorf("expected empty, got %q", got)
				}
			},
		},
		{
			name:        "non-JSON line passed through unchanged",
			input:       "plain text line",
			wantContain: []string{"plain text line"},
		},
		{
			name:  "multi-line: non-JSON and JSON mixed",
			input: "plain\n{\"level\":\"info\",\"msg\":\"hello\"}\nplain2",
			check: func(t *testing.T, got string) {
				t.Helper()
				parts := strings.Split(got, "\n")
				if len(parts) != 3 {
					t.Fatalf("expected 3 lines, got %d", len(parts))
				}
				if parts[0] != "plain" {
					t.Errorf("line 0: got %q, want %q", parts[0], "plain")
				}
				if parts[2] != "plain2" {
					t.Errorf("line 2: got %q, want %q", parts[2], "plain2")
				}
				if !strings.Contains(parts[1], "hello") {
					t.Errorf("line 1: expected msg 'hello' in %q", parts[1])
				}
			},
		},
		{
			name:        "known kind mcp uppercased, not in key=val",
			input:       `{"level":"info","msg":"ok","kind":"mcp","service":"svc"}`,
			wantContain: []string{"MCP"},
			wantAbsent:  []string{"kind="},
		},
		{
			name:        "known kind agent uppercased, not in key=val",
			input:       `{"level":"info","msg":"ok","kind":"agent","service":"svc"}`,
			wantContain: []string{"AGENT"},
			wantAbsent:  []string{"kind="},
		},
		{
			name:        "known kind daemon uppercased, not in key=val",
			input:       `{"level":"info","msg":"ok","kind":"daemon","service":"svc"}`,
			wantContain: []string{"DAEMON"},
			wantAbsent:  []string{"kind="},
		},
		{
			name:        "unknown kind resolved from svcKind map shows resolved kind and original as key=val",
			input:       `{"level":"info","msg":"indexed","kind":"local","service":"my-server"}`,
			svcKind:     map[string]string{"my-server": "mcp"},
			wantContain: []string{"MCP", "kind="},
		},
		{
			name:       "unknown kind with no map entry: kind not shown as key=val",
			input:      `{"level":"info","msg":"indexed","kind":"local","service":"unknown-svc"}`,
			wantAbsent: []string{"kind="},
		},
		{
			name:        "standard fields not in key=val, extra fields are",
			input:       `{"time":"2024-01-01T00:00:00Z","level":"info","msg":"hi","service":"svc","kind":"mcp","extra":"yes"}`,
			wantAbsent:  []string{"time=", "level=", "msg=", "service=", "kind="},
			wantContain: []string{"extra="},
		},
		{
			name:        "level error",
			input:       `{"level":"error","msg":"m"}`,
			wantContain: []string{"ERROR"},
		},
		{
			name:        "level warn",
			input:       `{"level":"warn","msg":"m"}`,
			wantContain: []string{"WARN"},
		},
		{
			name:        "level debug",
			input:       `{"level":"debug","msg":"m"}`,
			wantContain: []string{"DEBUG"},
		},
		{
			name:        "level info",
			input:       `{"level":"info","msg":"m"}`,
			wantContain: []string{"INFO"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := prettifyLogs(tc.input, tc.svcKind)
			for _, want := range tc.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q: %q", want, got)
				}
			}
			for _, absent := range tc.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("output should not contain %q: %q", absent, got)
				}
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}
