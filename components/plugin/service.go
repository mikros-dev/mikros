package plugin

import (
	"context"

	fenv "github.com/mikros-dev/mikros/apis/features/env"
	ferrors "github.com/mikros-dev/mikros/apis/features/errors"
	flogger "github.com/mikros-dev/mikros/apis/features/logger"
	mcontext "github.com/mikros-dev/mikros/components/context"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/service"
)

// Service is an internal package behavior that all supported service types must
// have.
type Service interface {
	// Name must return the implementation name. It's recommended to use a
	// kebab-case here.
	Name() string

	// Info should return some service informative fields to be logged while
	// the application is starting.
	Info() []flogger.Attribute

	// Initialize must be the place to initialize everything that needs information
	// from the framework.
	Initialize(ctx context.Context, opt *ServiceOptions) error

	// Run must put the server in execution. It can block or not the call.
	Run(ctx context.Context, srv interface{}) error

	// Stop should stop the service with a graceful shutdown.
	Stop(ctx context.Context) error
}

// ServiceSettings is an optional behavior that a plugin may have to load custom
// settings from the service 'service.toml' file.
type ServiceSettings interface {
	// Definitions must return the loaded service definitions from the
	// 'service.toml' file.
	//
	// To keep the framework standard, it's recommended that these custom
	// features settings use the service Name() as its main object inside
	// the TOML file. Like the example:
	//
	// [custom_service_name]
	//   custom_setting_a = 42
	//   custom_setting_b = "hello"
	//
	Definitions(path string) (definition.ExternalServiceEntry, error)
}

// ServiceOptions gathers all available options to create a service object.
type ServiceOptions struct {
	Port           service.ServerPort
	Type           definition.ServiceType
	Name           service.Name
	Product        string
	Logger         flogger.LoggerAPI
	Errors         ferrors.ErrorAPI
	ServiceContext *mcontext.ServiceContext
	Tags           map[string]string
	Service        options.ServiceOptions
	Definitions    *definition.Definitions
	Features       *FeatureSet
	ServiceHandler interface{}
	Env            fenv.EnvAPI
}
