package options

// Internal feature names
const (
	FeatureNamePrefix = "mikros_framework-"

	// Internal features

	HttpFeatureName       = FeatureNamePrefix + "http"
	LoggerFeatureName     = FeatureNamePrefix + "logger"
	ErrorsFeatureName     = FeatureNamePrefix + "errors"
	DefinitionFeatureName = FeatureNamePrefix + "definition"

	// These HTTP features plugins don't exist here, but to be supported by
	// internal services, they must have these names.

	HttpCorsFeatureName        = FeatureNamePrefix + "http_cors"
	HttpAuthFeatureName        = FeatureNamePrefix + "http_auth"
	TracingFeatureName         = FeatureNamePrefix + "tracing"
	TrackerFeatureName         = FeatureNamePrefix + "tracker"
	LoggerExtractorFeatureName = FeatureNamePrefix + "logger_extractor"
	PanicRecoveryFeatureName   = FeatureNamePrefix + "panic_recovery"
)
