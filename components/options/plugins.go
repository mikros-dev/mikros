package options

const (
	// PluginNamePrefix is the prefix for all features exported by the framework.
	PluginNamePrefix = "mikros_framework-"
)

// Internal features
const (
	HTTPFeatureName       = PluginNamePrefix + "http"
	LoggerFeatureName     = PluginNamePrefix + "logger"
	ErrorsFeatureName     = PluginNamePrefix + "errors"
	DefinitionFeatureName = PluginNamePrefix + "definition"
	EnvFeatureName        = PluginNamePrefix + "env"
)

// These HTTP features plugins don't exist here, but to be supported by
// internal runtimes, they must have these names.
const (
	HTTPCorsIntegrationName        = PluginNamePrefix + "http_cors"
	HTTPSpecAuthIntegrationName    = PluginNamePrefix + "http_spec_auth"
	HTTPAuthIntegrationName        = PluginNamePrefix + "http_auth"
	TracingIntegrationName         = PluginNamePrefix + "tracing"
	TrackerIntegrationName         = PluginNamePrefix + "tracker"
	LoggerExtractorIntegrationName = PluginNamePrefix + "logger_extractor"
	PanicRecoveryIntegrationName   = PluginNamePrefix + "panic_recovery"
)
