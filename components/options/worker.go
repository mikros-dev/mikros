package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

// WorkerServiceOptions represents configuration options specific to services
// of type worker.
type WorkerServiceOptions struct{}

// Kind returns the ServiceType associated with worker services as
// definition.ServiceTypeWorker.
func (n *WorkerServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceTypeWorker
}
