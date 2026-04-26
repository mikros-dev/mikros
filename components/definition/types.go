package definition

import (
	"strings"
)

type runtimeTypeCtx struct{}

// RuntimeType represents a supported type of runtime implemented by mikros.
type RuntimeType struct {
	name string
}

// Supported runtime types.
var (
	RuntimeTypeGRPC     = CreateRuntimeType("grpc")
	RuntimeTypeHTTPSpec = CreateRuntimeType("http-spec")
	RuntimeTypeHTTP   = CreateRuntimeType("http")
	RuntimeTypeWorker = CreateRuntimeType("worker")
	RuntimeTypeScript = CreateRuntimeType("script")
)

const (
	unknownType = "unknown"
)

// CreateRuntimeType creates a new RuntimeType with the specified name.
func CreateRuntimeType(name string) RuntimeType {
	return RuntimeType{name: name}
}

func (r RuntimeType) String() string {
	return r.name
}

// DeploymentEnv represents the deployment environment of a service.
type DeploymentEnv int32

// Supported environments.
const (
	DeploymentEnvUnknown DeploymentEnv = iota
	DeploymentEnvProduction
	DeploymentEnvTest
	DeploymentEnvDevelopment
	DeploymentEnvLocal
)

func (d *DeploymentEnv) String() string {
	switch *d {
	case DeploymentEnvProduction:
		return "prod"
	case DeploymentEnvTest:
		return "test"
	case DeploymentEnvDevelopment:
		return "dev"
	case DeploymentEnvLocal:
		return "local"
	default:
		return unknownType
	}
}

// FromString converts a string input to its corresponding DeploymentEnv
// enumeration value.
func (d *DeploymentEnv) FromString(in string) DeploymentEnv {
	switch in {
	case "prod":
		return DeploymentEnvProduction
	case "test":
		return DeploymentEnvTest
	case "dev":
		return DeploymentEnvDevelopment
	case "local":
		return DeploymentEnvLocal
	}

	return DeploymentEnvUnknown

}

// UnmarshalText converts a text input to a DeploymentEnv.
func (d *DeploymentEnv) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "prod":
		*d = DeploymentEnvProduction
	case "test":
		*d = DeploymentEnvTest
	case "dev":
		*d = DeploymentEnvDevelopment
	case "local":
		*d = DeploymentEnvLocal
	default:
		*d = DeploymentEnvUnknown
	}

	return nil
}

// SupportedRuntimeTypes gives a slice of all supported runtime types.
func SupportedRuntimeTypes() []string {
	var s []string
	types := []RuntimeType{
		RuntimeTypeGRPC,
		RuntimeTypeHTTPSpec,
		RuntimeTypeHTTP,
		RuntimeTypeWorker,
		RuntimeTypeScript,
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
