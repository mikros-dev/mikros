package plugin

import (
	"context"
	"fmt"
)

// FeatureSet gathers all features that a service can use during its execution.
type FeatureSet struct {
	features        map[string]*registeredFeature
	orderedFeatures []*registeredFeature
}

type FeatureSetIterator struct {
	index    int
	features []*registeredFeature
}

type registeredFeature struct {
	name         string
	feature      Feature
	dependencies []string
}

// NewFeatureSet creates a new FeatureSet.
func NewFeatureSet() *FeatureSet {
	return &FeatureSet{
		features: make(map[string]*registeredFeature),
	}
}

// InitializeAll initializes all previously registered feature (in the order
// they were registered).
func (s *FeatureSet) InitializeAll(ctx context.Context, options *InitializeOptions) error {
	for _, feature := range s.orderedFeatures {
		allowOptions := &CanBeInitializedOptions{
			DeploymentEnv: options.Env.DeploymentEnv(),
			Definitions:   options.Definitions,
		}

		createOptions := &InitializeOptions{
			Logger:          options.Logger,
			Errors:          options.Errors,
			Definitions:     options.Definitions,
			Tags:            options.Tags,
			ServiceContext:  options.ServiceContext,
			Dependencies:    s.getDependentFeatures(feature.dependencies),
			RunTimeFeatures: options.RunTimeFeatures,
			Env:             options.Env,
		}

		if err := s.initializeFeature(ctx, feature.feature, allowOptions, createOptions); err != nil {
			return err
		}
	}

	return nil
}

func (s *FeatureSet) getDependentFeatures(names []string) map[string]Feature {
	deps := make(map[string]Feature)
	for _, name := range names {
		if f, ok := s.features[name]; ok {
			deps[name] = f.feature
		}
	}

	return deps
}

func (s *FeatureSet) initializeFeature(ctx context.Context, feature Feature, allow *CanBeInitializedOptions, create *InitializeOptions) error {
	enabled := feature.CanBeInitialized(allow)
	feature.UpdateInfo(UpdateInfoEntry{Enabled: enabled, Logger: create.Logger, Errors: create.Errors})

	if enabled {
		if err := feature.Initialize(ctx, create); err != nil {
			return err
		}
	}

	return nil
}

// Register registers an internal feature that will be initialized, if
// allowed, to be used by a service. The features will be initialized in the
// order they are registered.
func (s *FeatureSet) Register(name string, feature Feature, dependencies ...string) {
	if feature != nil {
		// Gives the feature access to its name from this point on.
		feature.UpdateInfo(UpdateInfoEntry{Name: name})
		f := &registeredFeature{
			dependencies: dependencies,
			feature:      feature,
			name:         name,
		}

		s.features[name] = f
		s.orderedFeatures = append(s.orderedFeatures, f)
	}
}

func (s *FeatureSet) Feature(name string) (Feature, error) {
	feature, ok := s.features[name]
	if !ok {
		return nil, fmt.Errorf("could not find feature '%v'", name)
	}

	return feature.feature, nil
}

func (s *FeatureSet) Iterator() *FeatureSetIterator {
	return &FeatureSetIterator{
		features: s.orderedFeatures,
		index:    0,
	}
}

func (s *FeatureSet) Count() int {
	return len(s.features)
}

func (s *FeatureSet) Append(features *FeatureSet) {
	if features != nil {
		for _, feature := range features.orderedFeatures {
			s.features[feature.name] = feature
			s.orderedFeatures = append(s.orderedFeatures, feature)
		}
	}
}

func (s *FeatureSet) StartAll(ctx context.Context, srv interface{}) error {
	for _, feature := range s.features {
		if p, ok := feature.feature.(FeatureController); ok {
			if err := p.Start(ctx, srv); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *FeatureSet) CleanupAll(ctx context.Context) error {
	for _, feature := range s.features {
		if p, ok := feature.feature.(FeatureController); ok {
			if err := p.Cleanup(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *FeatureSetIterator) Next() (Feature, bool) {
	if i.index < len(i.features) {
		e := i.features[i.index]
		i.index++
		return e.feature, true
	}

	return nil, false
}

func (i *FeatureSetIterator) Reset() {
	i.index = 0
}
