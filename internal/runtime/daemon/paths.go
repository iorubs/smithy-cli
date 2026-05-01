// Package daemon implements `smithy stack up` background mode.
//
// Stacks are addressed by a user-supplied name. All artefacts for a
// stack live under `~/.smithy/stacks/<name>/` so multiple stacks
// coexist on one host and lifecycle commands can address them by
// name without needing the original config path.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Paths is the set of on-disk artefacts a single backgrounded stack owns.
type Paths struct {
	// Dir is the per-stack directory that holds the socket, daemon log, and metadata file.
	Dir string
	// Socket is the absolute path to the Unix domain socket the daemon serves /status on.
	Socket string
	// DaemonLog is the absolute path to the daemon process's own log file (slog output).
	DaemonLog string
	// Meta is the absolute path to the stack metadata file. It holds
	// the daemon PID, original config path, and start timestamp; readers should use ReadMeta.
	Meta string
}

// nameRE constrains stack names to a portable, shell-safe subset.
var nameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// ValidateName returns nil if name is a legal stack identifier.
func ValidateName(name string) error {
	if name == "" {
		return errors.New("stack name is required")
	}
	if !nameRE.MatchString(name) {
		return fmt.Errorf("invalid stack name %q: must match %s", name, nameRE.String())
	}
	return nil
}

// DeriveName produces a stack name from a stack file path by
// stripping the directory and extension. Returns an error if the
// derived name does not satisfy ValidateName.
func DeriveName(stackPath string) (string, error) {
	base := filepath.Base(stackPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if err := ValidateName(stem); err != nil {
		return "", fmt.Errorf("derive name from %q: %w", stackPath, err)
	}
	return stem, nil
}

// Root returns the directory holding all stack subdirectories for the
// current working directory (`<cwd>/.smithy/`). Lifecycle commands
// only see stacks started from the same directory.
func Root() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("daemon: getwd: %w", err)
	}
	return filepath.Join(cwd, ".smithy"), nil
}

// PathsFor returns the artefact paths for the given stack name. The
// directory is not created here; callers that need it on disk
// (Run, SpawnDetached) MkdirAll explicitly.
func PathsFor(name string) (Paths, error) {
	if err := ValidateName(name); err != nil {
		return Paths{}, err
	}
	root, err := Root()
	if err != nil {
		return Paths{}, err
	}
	dir := filepath.Join(root, name)
	return Paths{
		Dir:       dir,
		Socket:    filepath.Join(dir, "daemon.sock"),
		DaemonLog: filepath.Join(dir, "daemon.log"),
		Meta:      filepath.Join(dir, "stack.json"),
	}, nil
}

// ListNames returns every directory name under Root, regardless of
// whether the underlying daemon is alive. Callers that care about
// liveness should probe each socket via IsStackRunning.
func ListNames() ([]string, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("daemon: read %s: %w", root, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if nameRE.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
