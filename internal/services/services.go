package services

import (
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/internal/services/grpc"
	"github.com/mikros-dev/mikros/internal/services/http"
	"github.com/mikros-dev/mikros/internal/services/http_spec"
	"github.com/mikros-dev/mikros/internal/services/script"
	"github.com/mikros-dev/mikros/internal/services/worker"
)

func Services() *plugin.ServiceSet {
	services := plugin.NewServiceSet()

	services.Register(grpc.New())
	services.Register(http_spec.New())
	services.Register(http.New())
	services.Register(worker.New())
	services.Register(script.New())

	return services
}
