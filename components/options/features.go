package options

// Internal feature names
const (
	FeatureNamePrefix = "mikros_framework-"

	// Internal features

	HTTPFeatureName       = FeatureNamePrefix + "http"
	LoggerFeatureName     = FeatureNamePrefix + "logger"
	ErrorsFeatureName     = FeatureNamePrefix + "errors"
	DefinitionFeatureName = FeatureNamePrefix + "definition"
	EnvFeatureName        = FeatureNamePrefix + "env"

	// These HTTP features plugins don't exist here, but to be supported by
	// internal services, they must have these names.

	HTTPCorsFeatureName        = FeatureNamePrefix + "http_cors"
	HTTPSpecAuthFeatureName    = FeatureNamePrefix + "http_spec_auth"
	HTTPAuthFeatureName        = FeatureNamePrefix + "http_auth"
	TracingFeatureName         = FeatureNamePrefix + "tracing"
	TrackerFeatureName         = FeatureNamePrefix + "tracker"
	LoggerExtractorFeatureName = FeatureNamePrefix + "logger_extractor"
	PanicRecoveryFeatureName   = FeatureNamePrefix + "panic_recovery"
)
