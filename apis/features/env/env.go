package env

import (
	"github.com/mikros-dev/mikros/components/definition"
)

type EnvAPI interface {
	// Get searches and returns the value of an environment variable in string
	// format.
	Get(name string) string

	// GetInt searches and returns the value of an environment variable in
	// an int format.
	GetInt(name string) (int, error)

	// GetBool searches and returns the value of an environment variable in
	// a boolean format.
	GetBool(name string) (bool, error)

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
