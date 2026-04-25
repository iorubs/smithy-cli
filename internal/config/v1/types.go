// Package v1 defines the v1 schema for smithy-compose.yaml.
package v1

// Version is the schema version this package handles.
const Version = "1"

// Config is the root of smithy-compose.yaml.
type Config struct {
	// Schema version. Must be "1".
	Version string `yaml:"version" smithy:"required"`
}
