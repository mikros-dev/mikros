package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

// ScriptServiceOptions represents configuration options specific to script-based services.
type ScriptServiceOptions struct{}

// Kind returns the service type corresponding to a script-based service as
// definition.ServiceTypeScript.
func (s *ScriptServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceTypeScript
}
