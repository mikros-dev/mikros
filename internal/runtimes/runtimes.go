package runtimes

import (
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/internal/runtimes/grpc"
	"github.com/mikros-dev/mikros/internal/runtimes/http"
	"github.com/mikros-dev/mikros/internal/runtimes/http_spec"
	"github.com/mikros-dev/mikros/internal/runtimes/script"
	"github.com/mikros-dev/mikros/internal/runtimes/worker"
)

// Runtimes returns a RuntimeSet with all the supported runtimes by mikros.
func Runtimes() *plugin.RuntimeSet {
	set := plugin.NewRuntimeSet()

	set.Register(grpc.New())
	set.Register(http_spec.New())
	set.Register(http.New())
	set.Register(worker.New())
	set.Register(script.New())

	return set
}
