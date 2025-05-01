package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

type ScriptServiceOptions struct{}

func (s *ScriptServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_Script
}
