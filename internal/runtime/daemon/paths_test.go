package daemon

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	ok := []string{"sample", "a", "0", "sample-prod", "api_v2", "abc123"}
	for _, n := range ok {
		if err := ValidateName(n); err != nil {
			t.Errorf("ValidateName(%q) unexpected error: %v", n, err)
		}
	}
	bad := []string{"", "-leading", "_leading", "Upper", "with space", "with/slash", "with.dot", strings.Repeat("a", 65)}
	for _, n := range bad {
		if err := ValidateName(n); err == nil {
			t.Errorf("ValidateName(%q) should have errored", n)
		}
	}
}

func TestPathsFor(t *testing.T) {
	p, err := PathsFor("sample")
	if err != nil {
		t.Fatalf("PathsFor: %v", err)
	}
	if !filepath.IsAbs(p.Dir) {
		t.Errorf("Dir not absolute: %q", p.Dir)
	}
	if filepath.Base(p.Dir) != "sample" {
		t.Errorf("Dir basename = %q, want %q", filepath.Base(p.Dir), "sample")
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
}

func TestPathsForRejectsInvalid(t *testing.T) {
	if _, err := PathsFor("Bad Name"); err == nil {
		t.Errorf("expected error for invalid name")
	}
}
