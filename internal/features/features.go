package features

import (
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/internal/features/definition"
	"github.com/mikros-dev/mikros/internal/features/env"
	"github.com/mikros-dev/mikros/internal/features/errors"
	"github.com/mikros-dev/mikros/internal/features/http"
	"github.com/mikros-dev/mikros/internal/features/logger"
)

// Features returns the set of features that are available in mikros.
func Features() *plugin.FeatureSet {
	features := plugin.NewFeatureSet()

	features.Register(options.HTTPFeatureName, http.New())
	features.Register(options.LoggerFeatureName, logger.New())
	features.Register(options.ErrorsFeatureName, errors.New())
	features.Register(options.DefinitionFeatureName, definition.New())
	features.Register(options.EnvFeatureName, env.New())

	return features
}
