package definition

// API provides access to service metadata loaded from the service.toml file.
//
// This interface is implemented by the mikros framework and made available to
// services that opt into the "definition" feature. It allows services to retrieve
// identifying information, such as the service name, without needing to manage
// configuration parsing directly.
type API interface {
	// ServiceName returns the name of the service as defined in the service.toml file.
	ServiceName() string
}
