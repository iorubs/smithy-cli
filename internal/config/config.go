// Package config loads and version-routes smithy-stack.yaml configs.
// Each schema version lives in its own sub-package and satisfies [VersionSchema].
package config

import (
	"fmt"
	"maps"
	"slices"

	v1 "github.com/iorubs/smithy-cli/internal/config/v1"
	"go.yaml.in/yaml/v4"
)

// Versions is the single source of truth for which schema versions are
// accepted. Each entry satisfies [VersionSchema].
var Versions = map[string]VersionSchema{
	v1.Version: v1.Schema{},
}

// VersionSchema is the contract each config version must satisfy.
// The Parse method must return the latest Config type (converting if needed).
type VersionSchema interface {
	Parse([]byte) (*Config, error)
	RootType() any
	TypesSources() []string
}

// TypesSources returns the raw Go source files for the latest version's types.
var TypesSources = v1.TypesSources

// Type aliases: always point to the latest version.
type (
	Config = v1.Config
)

// Parse parses raw YAML bytes, detects the version, delegates to the
// correct versioned parser, and converts the result to the latest
// Config type.
func Parse(data []byte) (*Config, error) {
	var header struct {
		Version string `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	e, ok := Versions[header.Version]
	if !ok {
		return nil, fmt.Errorf("unsupported config version %q (supported: %s)",
			header.Version, slices.Sorted(maps.Keys(Versions)))
	}
	return e.Parse(data)
}
