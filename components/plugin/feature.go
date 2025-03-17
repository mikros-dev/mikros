package plugin

import (
	"context"

	errorsApi "github.com/mikros-dev/mikros/apis/errors"
	loggerApi "github.com/mikros-dev/mikros/apis/logger"
	mcontext "github.com/mikros-dev/mikros/components/context"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
	"github.com/mikros-dev/mikros/components/testing"
)

// Feature is a set of methods that all framework feature, internal or external,
// must implement to be supported.
type Feature interface {
	// CanBeInitialized is the method executed to check if the feature is
	// allowed to be used by the current service or not. Here the feature
	// should check everything that is needs to return this information.
	CanBeInitialized(options *CanBeInitializedOptions) bool

	// Initialize is the method that "creates" the feature, where all its
	// required initialization must be made.
	Initialize(ctx context.Context, options *InitializeOptions) error

	// Fields should return informative fields to be logged at the beginning
	// of the execution.
	Fields() []loggerApi.Attribute

	// FeatureEntry is a set of methods that must provide information related
	// to the feature itself.
	FeatureEntry
}

// FeatureEntry is a set of methods that provide information related to the
// feature.
type FeatureEntry interface {
	// UpdateInfo is an internal method that allows a feature to have its
	// information, such as its name, if it's enabled or not, internally.
	UpdateInfo(info UpdateInfoEntry)

	// IsEnabled returns true or false if the feature is currently enabled or not.
	IsEnabled() bool

	// Name returns the feature name.
	Name() string
}

// UpdateInfoEntry is a structure used to update internal FeatureEntry types
// according its initialized members.
type UpdateInfoEntry struct {
	Enabled bool
	Name    string
	Logger  loggerApi.Logger
	Errors  errorsApi.ErrorFactory
}

// FeatureController is an optional behavior that a feature may have if it needs
// to execute tasks with the service main object.
type FeatureController interface {
	// Start is a method where the plugin receives the service main object to
	// initialize custom tasks.
	Start(ctx context.Context, srv interface{}) error

	// Cleanup should free all resources allocated by the plugin or stop any
	// internal process.
	Cleanup(ctx context.Context) error
}

// FeatureSettings is an optional behavior that a feature may have to load custom
// settings from the service 'service.toml' file.
type FeatureSettings interface {
	// Definitions must return the feature definitions loaded from the
	// 'service.toml' file.
	//
	// To keep the framework standard, it's recommended that these custom
	// features settings reside inside a 'features' object inside the TOML
	// file. Like the example:
	//
	// [features.custom]
	//   custom_setting_a = 42
	//   custom_setting_b = "hello"
	//
	Definitions(path string) (definition.ExternalFeatureEntry, error)
}

// FeatureExternalAPI is a behavior that every external feature must have so that
// their API can be used from services. This is specific for features that support
// test mocking.
type FeatureExternalAPI interface {
	ServiceAPI() interface{}
}

// FeatureInternalAPI is a behaviour that a feature can have to provide an API
// to be used inside the framework or its extensions.
type FeatureInternalAPI interface {
	FrameworkAPI() interface{}
}

// FeatureTester is a behavior that a feature should implement to be mocked
// in a unit test.
type FeatureTester interface {
	// Setup is responsible for changing internal behaviors when running a
	// specific unit test.
	Setup(ctx context.Context, t *testing.Testing)

	// Teardown is responsible for cleaning up all resources allocated when
	// Setup was called before. It's important to notice here that after this
	// call a new call to the Setup API must be made to run a new test.
	Teardown(ctx context.Context, t *testing.Testing)

	// DoTest is where the feature executes its specific tests previously
	// adjusted in the testing.Testing.Options.FeatureOptions.
	DoTest(ctx context.Context, t *testing.Testing, serviceName service.Name) error
}

// CanBeInitializedOptions gathers all information passed to the CanBeInitialized
// method of a Feature interface.
type CanBeInitializedOptions struct {
	DeploymentEnv definition.ServiceDeploy
	Definitions   *definition.Definitions
}

// InitializeOptions gathers all information passed to the Initialize method of
// a Feature interface, allowing a feature to be properly initialized.
type InitializeOptions struct {
	Logger          loggerApi.Logger
	Errors          errorsApi.ErrorFactory
	Definitions     *definition.Definitions
	Tags            map[string]string
	ServiceContext  *mcontext.ServiceContext
	Dependencies    map[string]Feature
	RunTimeFeatures map[string]interface{}
	Env             Env
}
