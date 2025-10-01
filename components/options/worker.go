package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

type WorkerServiceOptions struct{}

func (n *WorkerServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_Worker
}
