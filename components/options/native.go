package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

type NativeServiceOptions struct{}

func (n *NativeServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_Native
}
