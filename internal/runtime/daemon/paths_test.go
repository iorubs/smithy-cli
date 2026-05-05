package daemon

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple word", "sample", false},
		{"single char", "a", false},
		{"single digit", "0", false},
		{"hyphenated", "sample-prod", false},
		{"underscore", "api_v2", false},
		{"alphanumeric", "abc123", false},
		{"empty", "", true},
		{"leading hyphen", "-leading", true},
		{"leading underscore", "_leading", true},
		{"uppercase", "Upper", true},
		{"space", "with space", true},
		{"slash", "with/slash", true},
		{"dot", "with.dot", true},
		{"too long", strings.Repeat("a", 65), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v; wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestPathsFor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "sample", false},
		{"invalid name", "Bad Name", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := PathsFor(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("PathsFor: %v", err)
			}
			if !filepath.IsAbs(p.Dir) {
				t.Errorf("Dir not absolute: %q", p.Dir)
			}
			if filepath.Base(p.Dir) != tt.input {
				t.Errorf("Dir basename = %q, want %q", filepath.Base(p.Dir), tt.input)
			}
			if filepath.Base(p.Socket) != "daemon.sock" {
				t.Errorf("Socket basename = %q", filepath.Base(p.Socket))
			}
			if filepath.Base(p.DaemonLog) != "daemon.log" {
				t.Errorf("DaemonLog basename = %q", filepath.Base(p.DaemonLog))
			}
			if filepath.Base(p.Meta) != "stack.json" {
				t.Errorf("Meta basename = %q", filepath.Base(p.Meta))
			}
		})
	}
}
