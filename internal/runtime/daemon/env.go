package daemon

import (
	"fmt"
	"path/filepath"

	"github.com/iorubs/smithy-cli/internal/config"
	"github.com/iorubs/smithy-cli/internal/runtime/dotenv"
)

// loadStackEnv applies env_file declarations from the stack config
// to the daemon process environment, plus a default
// `<stackdir>/.env` when no env_file is declared. Compose-style
// precedence: shell wins, then `env_file:` in declared order, then
// default `.env`. Each layer skips keys already set by a
// higher-priority layer (Load does "set if not present").
func loadStackEnv(stackPath string, cfg *config.Config) error {
	if cfg == nil {
		return nil
	}
	stackDir, err := filepath.Abs(filepath.Dir(stackPath))
	if err != nil {
		return fmt.Errorf("resolve stack dir: %w", err)
	}
	for _, ef := range cfg.EnvFile {
		path := ef
		if !filepath.IsAbs(path) {
			path = filepath.Join(stackDir, path)
		}
		if err := dotenv.Load(path); err != nil {
			return err
		}
	}
	if len(cfg.EnvFile) == 0 {
		if err := dotenv.Load(filepath.Join(stackDir, ".env")); err != nil {
			return err
		}
	}
	return nil
}
