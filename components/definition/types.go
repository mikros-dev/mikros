package definition

type serviceTypeCtx struct{}

type ServiceType struct {
	name string
}

var (
	ServiceType_gRPC   = CreateServiceType("grpc")
	ServiceType_HTTP   = CreateServiceType("http")
	ServiceType_Native = CreateServiceType("native")
)

const (
	unknownType = "unknown"
)

func CreateServiceType(name string) ServiceType {
	return ServiceType{name: name}
}

func (s ServiceType) String() string {
	return s.name
}

type ServiceDeploy int32

const (
	ServiceDeploy_Unknown ServiceDeploy = iota
	ServiceDeploy_Production
	ServiceDeploy_Test
	ServiceDeploy_Development
	ServiceDeploy_Local
)

func (e ServiceDeploy) String() string {
	switch e {
	case ServiceDeploy_Production:
		return "prod"
	case ServiceDeploy_Test:
		return "test"
	case ServiceDeploy_Development:
		return "dev"
	case ServiceDeploy_Local:
		return "local"
	}

	return unknownType
}

func (e ServiceDeploy) FromString(in string) ServiceDeploy {
	switch in {
	case "prod":
		return ServiceDeploy_Production
	case "test":
		return ServiceDeploy_Test
	case "dev":
		return ServiceDeploy_Development
	case "local":
		return ServiceDeploy_Local
	}

	return ServiceDeploy_Unknown
}

// SupportedServiceTypes gives a slice of all supported service types.
func SupportedServiceTypes() []string {
	var s []string
	types := []ServiceType{
		ServiceType_gRPC,
		ServiceType_HTTP,
		ServiceType_Native,
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
	// Enabled should enable or disable the feature. Default is always to start
	// with the feature disabled.
	Enabled bool `toml:"enabled,omitempty"`
}

func (f FeatureEntry) IsEnabled() bool {
	return f.Enabled
}
