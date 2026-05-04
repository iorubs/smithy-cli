package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	body := "" +
		"# comment\n" +
		"FOO=bar\n" +
		"  QUOTED=\"with spaces\"  # trailing comment\n" +
		"SINGLE='hashed#inside'\n" +
		"\n" +
		"EMPTY=\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	keys := []string{"FOO", "QUOTED", "SINGLE", "EMPTY"}
	for _, k := range keys {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	t.Cleanup(func() {
		for _, k := range keys {
			os.Unsetenv(k)
		}
	})

	tests := []struct {
		key  string
		want string
	}{
		{"FOO", "bar"},
		{"QUOTED", "with spaces"},
		{"SINGLE", "hashed#inside"},
		{"EMPTY", ""},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := os.Getenv(tt.key); got != tt.want {
				t.Errorf("%s = %q; want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestLoadEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		envKey  string
		envVal  string
		wantErr bool
		check   func(t *testing.T)
	}{
		{
			name:    "shell env wins over file",
			content: "KEEP=fromfile\n",
			envKey:  "KEEP",
			envVal:  "fromshell",
			check: func(t *testing.T) {
				if got := os.Getenv("KEEP"); got != "fromshell" {
					t.Fatalf("KEEP = %q; want shell value preserved", got)
				}
			},
		},
		{
			name:    "invalid line errors",
			content: "not-a-pair\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, ".env")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envVal)
			}
			err := Load(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestLoadMissingFileNotAnError(t *testing.T) {
	if err := Load(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("Load missing: %v", err)
	}
}
