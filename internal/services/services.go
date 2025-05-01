package services

import (
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/internal/services/grpc"
	"github.com/mikros-dev/mikros/internal/services/http"
	"github.com/mikros-dev/mikros/internal/services/native"
	"github.com/mikros-dev/mikros/internal/services/script"
)

func Services() *plugin.ServiceSet {
	services := plugin.NewServiceSet()

	services.Register(grpc.New())
	services.Register(http.New())
	services.Register(native.New())
	services.Register(script.New())

	return services
}
