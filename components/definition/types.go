package definition

import (
	"strings"
)

type serviceTypeCtx struct{}

// ServiceType represents a supported type of service implemented internally.
type ServiceType struct {
	name string
}

// Supported service types.
var (
	ServiceTypeGRPC     = CreateServiceType("grpc")
	ServiceTypeHTTPSpec = CreateServiceType("http-spec")
	ServiceTypeHTTP     = CreateServiceType("http")
	ServiceTypeWorker   = CreateServiceType("worker")
	ServiceTypeScript   = CreateServiceType("script")
)

const (
	unknownType = "unknown"
)

// CreateServiceType creates a new ServiceType with the specified name.
func CreateServiceType(name string) ServiceType {
	return ServiceType{name: name}
}

func (s ServiceType) String() string {
	return s.name
}

// ServiceDeploy represents the deployment environment of a service.
type ServiceDeploy int32

// Supported environments.
const (
	ServiceDeployUnknown ServiceDeploy = iota
	ServiceDeployProduction
	ServiceDeployTest
	ServiceDeployDevelopment
	ServiceDeployLocal
)

func (e ServiceDeploy) String() string {
	switch e {
	case ServiceDeployProduction:
		return "prod"
	case ServiceDeployTest:
		return "test"
	case ServiceDeployDevelopment:
		return "dev"
	case ServiceDeployLocal:
		return "local"
	default:
		return unknownType
	}
}

// FromString converts a string input to its corresponding ServiceDeploy
// enumeration value.
func (e ServiceDeploy) FromString(in string) ServiceDeploy {
	switch in {
	case "prod":
		return ServiceDeployProduction
	case "test":
		return ServiceDeployTest
	case "dev":
		return ServiceDeployDevelopment
	case "local":
		return ServiceDeployLocal
	}

	return ServiceDeployUnknown

}

// UnmarshalText converts a text input to a ServiceDeploy.
func (e *ServiceDeploy) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "prod":
		*e = ServiceDeployProduction
	case "test":
		*e = ServiceDeployTest
	case "dev":
		*e = ServiceDeployDevelopment
	case "local":
		*e = ServiceDeployLocal
	default:
		*e = ServiceDeployUnknown
	}

	return nil
}

// SupportedServiceTypes gives a slice of all supported service types.
func SupportedServiceTypes() []string {
	var s []string
	types := []ServiceType{
		ServiceTypeGRPC,
		ServiceTypeHTTPSpec,
		ServiceTypeHTTP,
		ServiceTypeWorker,
		ServiceTypeScript,
	}

	for _, t := range types {
		s = append(s, t.String())
	}

	return s
}

// SupportedLanguages gives a slice of supported programming languages.
func SupportedLanguages() []string {
	return []string{"go", "rust"}
}

// FeatureEntry is a structure that an external feature should have so that all
// presents at least a few common settings.
type FeatureEntry struct {
	// Enabled should enable or disable the feature. The default is always to
	// start with the feature disabled.
	Enabled bool `toml:"enabled,omitempty"`
}

// IsEnabled checks whether the feature is enabled or not.
func (f FeatureEntry) IsEnabled() bool {
	return f.Enabled
}

// Validator defines an interface for definitions that can self-validate.
type Validator interface {
	Validate() error
}
