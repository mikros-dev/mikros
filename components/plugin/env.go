package plugin

import (
	"github.com/mikros-dev/mikros/components/definition"
)

// Env is the plugin interface that allow plugins (features and services) to "access"
// the framework env.
type Env interface {
	// Get retrieves an environment variable value that was declared inside the
	// service.toml file.
	Get(key string) interface{}

	// DeploymentEnv gets the current service deployment environment.
	DeploymentEnv() definition.ServiceDeploy

	// TrackerHeaderName gives the current header name that contains the service
	// tracker ID (for HTTP services).
	TrackerHeaderName() string

	// IsCICD gets if the CI/CD is being running or not.
	IsCICD() bool

	// CoupledNamespace returns the namespace used by the services.
	CoupledNamespace() string

	// CoupledPort returns the port used to couple between services.
	CoupledPort() int32

	// GrpcPort returns the port number that gRPC services should use.
	GrpcPort() int32

	// HttpPort returns the port number that HTTP services should use.
	HttpPort() int32
}
