package env

import (
	"github.com/mikros-dev/mikros/components/definition"
)

// API provides access to environment-related settings for the running
// service.
//
// This interface is implemented by the mikros framework and is available to
// services that opt into the feature. It offers access to environment variables
// and deployment configuration values commonly required at runtime.
type API interface {
	// Get retrieves the value of an environment variable as a string.
	Get(name string) string

	// GetInt retrieves the value of an environment variable and parses it as
	// an integer.
	GetInt(name string) (int, error)

	// GetBool retrieves the value of an environment variable and parses it as
	// a boolean.
	GetBool(name string) (bool, error)

	// DeploymentEnv returns the current service deployment environment.
	DeploymentEnv() definition.ServiceDeploy

	// TrackerHeaderName returns the header name used to carry the service
	// tracker ID in HTTP requests and requests between services.
	TrackerHeaderName() string

	// IsCICD indicates whether the service is running in a CI/CD environment.
	IsCICD() bool

	// CoupledNamespace returns the namespace used for inter-service coupling.
	CoupledNamespace() string

	// CoupledPort returns the port number used for inter-service coupling.
	CoupledPort() int32

	// GrpcPort returns the port number to be used for gRPC services.
	GrpcPort() int32

	// HTTPPort returns the port number to be used for HTTP services.
	HTTPPort() int32
}
