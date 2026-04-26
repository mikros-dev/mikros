package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

// ScriptServiceOptions represents configuration options specific to script-based
// services.
type ScriptServiceOptions struct{}

// Kind returns the runtime type corresponding to a script-based service as
// definition.RuntimeTypeScript.
func (s *ScriptServiceOptions) Kind() definition.RuntimeType {
	return definition.RuntimeTypeScript
}
