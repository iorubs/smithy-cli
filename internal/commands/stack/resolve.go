package stack

import (
	"github.com/iorubs/smithy-cli/internal/runtime/daemon"
)

// ResolveStackName picks the stack name from a positional name or the
// stack file path. When name is given it is validated and returned;
// otherwise the name is derived from the config file stem (Kong has
// already applied the default "smithy-stack.yaml" via ConfigFlag).
func ResolveStackName(name, config string) (string, error) {
	if name != "" {
		if err := daemon.ValidateName(name); err != nil {
			return "", err
		}
		return name, nil
	}
	return daemon.DeriveName(config)
}
