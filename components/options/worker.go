package options

import (
	"github.com/mikros-dev/mikros/components/definition"
)

// WorkerServiceOptions represents configuration options specific to services
// of type worker.
type WorkerServiceOptions struct{}

// Kind returns the RuntimeType associated with worker services as
// definition.RuntimeTypeWorker.
func (n *WorkerServiceOptions) Kind() definition.RuntimeType {
	return definition.RuntimeTypeWorker
}
