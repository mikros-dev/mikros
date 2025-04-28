package features

import (
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/internal/features/definition"
	"github.com/mikros-dev/mikros/internal/features/errors"
	"github.com/mikros-dev/mikros/internal/features/http"
	"github.com/mikros-dev/mikros/internal/features/logger"
)

func Features() *plugin.FeatureSet {
	features := plugin.NewFeatureSet()

	features.Register(options.HttpFeatureName, http.New())
	features.Register(options.LoggerFeatureName, logger.New())
	features.Register(options.ErrorsFeatureName, errors.New())
	features.Register(options.DefinitionFeatureName, definition.New())

	return features
}
