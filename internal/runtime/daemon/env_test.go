package daemon

import (
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/iorubs/smithy-cli/internal/config/v1"
)

func TestLoadStackEnv(t *testing.T) {
	type fileSpec struct {
		name, body string
	}

	tests := []struct {
		name    string
		files   []fileSpec
		envFile []string
		setenv  map[string]string
		unset   []string
		want    map[string]string
	}{
		{
			name: "default .env loaded when env_file empty",
			files: []fileSpec{
				{".env", "DEFAULT_KEY=fromdefault\n"},
			},
			unset: []string{"DEFAULT_KEY"},
			want:  map[string]string{"DEFAULT_KEY": "fromdefault"},
		},
		{
			name: "explicit env_file skips default .env",
			files: []fileSpec{
				{".env", "FROM_DEFAULT=yes\n"},
				{"explicit.env", "FROM_EXPLICIT=yes\n"},
			},
			envFile: []string{"explicit.env"},
			unset:   []string{"FROM_DEFAULT", "FROM_EXPLICIT"},
			want: map[string]string{
				"FROM_EXPLICIT": "yes",
				"FROM_DEFAULT":  "",
			},
		},
		{
			name: "first declared file wins on conflict",
			files: []fileSpec{
				{"first.env", "PICK=fromfirst\n"},
				{"second.env", "PICK=fromsecond\n"},
			},
			envFile: []string{"first.env", "second.env"},
			unset:   []string{"PICK"},
			want:    map[string]string{"PICK": "fromfirst"},
		},
		{
			name: "shell value wins over file value",
			files: []fileSpec{
				{".env", "SHELL_WIN=fromfile\n"},
			},
			setenv: map[string]string{"SHELL_WIN": "fromshell"},
			want:   map[string]string{"SHELL_WIN": "fromshell"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			stack := filepath.Join(dir, "smithy-stack.yaml")
			if err := os.WriteFile(stack, []byte("version: \"1\"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			for _, k := range tt.unset {
				os.Unsetenv(k)
				t.Cleanup(func() { os.Unsetenv(k) })
			}
			for k, v := range tt.setenv {
				t.Setenv(k, v)
			}

			cfg := &v1.Config{Version: "1", EnvFile: tt.envFile}
			if err := loadStackEnv(stack, cfg); err != nil {
				t.Fatalf("loadStackEnv: %v", err)
			}
			for k, want := range tt.want {
				if got := os.Getenv(k); got != want {
					t.Errorf("%s = %q; want %q", k, got, want)
				}
			}
		})
	}
}
