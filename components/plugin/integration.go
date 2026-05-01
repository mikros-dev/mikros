package plugin

import (
	"context"
)

// Integration is a set of methods that all framework integrations, internal or
// external, must implement to be supported.
type Integration interface {
	// CanBeInitialized is the method executed to check if the integration is
	// allowed to be used by the current service or not. Here the integration
	// should check everything that is needs to return this information.
	CanBeInitialized(options *CanBeInitializedOptions) bool

	// Initialize is the method that "creates" the integration, where all its
	// required initialization must be made.
	Initialize(ctx context.Context, options *InitializeOptions) error

	// API returns the integration API.
	API() interface{}

	// IntegrationEntry is a set of methods that must provide information
	// related to the integration itself.
	IntegrationEntry
}

// IntegrationEntry is a set of methods that provide information related to the
// integration.
type IntegrationEntry interface {
	// UpdateInfo is an internal method that allows an integration to have its
	// information, such as its name, if it's enabled or not, internally.
	UpdateInfo(info UpdateInfoEntry)

	// IsEnabled returns true or false if the integration is currently enabled or
	// not.
	IsEnabled() bool

	// Name returns the integration name.
	Name() string
}

// IntegrationController is an optional behavior that an integration may have if
// it needs to execute tasks with the service main object.
type IntegrationController interface {
	// Start is a method where the plugin receives the service main object to
	// initialize custom tasks.
	Start(ctx context.Context, srv interface{}) error

	// Cleanup should free all resources allocated by the plugin or stop any
	// internal process.
	Cleanup(ctx context.Context) error
}
