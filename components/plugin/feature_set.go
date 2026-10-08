package plugin

import (
	"context"
	"fmt"
	"reflect"
)

// FeatureSet gathers all features that a service can use during its execution.
type FeatureSet struct {
	features         map[string]*registeredFeature
	orderedFeatures  []*registeredFeature
	declaredFeatures map[string]Feature
}

// FeatureSetIterator provides iteration capabilities over a collection of
// registered features in a FeatureSet.
type FeatureSetIterator struct {
	index    int
	features []Feature
}

type registeredFeature struct {
	name         string
	feature      Feature
	dependencies []string
	instances    []Feature
}

// NewFeatureSet creates a new FeatureSet.
func NewFeatureSet() *FeatureSet {
	return &FeatureSet{
		features:         make(map[string]*registeredFeature),
		declaredFeatures: make(map[string]Feature),
	}
}

// InitializeAll initializes all previously registered features (in the order
// they were registered).
func (s *FeatureSet) InitializeAll(ctx context.Context, options *InitializeOptions) error {
	return s.InitializeAllWithDeclarations(ctx, options, nil)
}

// InitializeAllWithDeclarations initializes registered singleton features and
// creates one independent instance per declaration for features that implement
// FeatureInstantiator.
func (s *FeatureSet) InitializeAllWithDeclarations(
	ctx context.Context,
	options *InitializeOptions,
	declarations []FeatureDeclaration,
) error {
	s.declaredFeatures = make(map[string]Feature)

	claimedAPIs := make(map[reflect.Type]string)
	for _, entry := range s.orderedFeatures {
		entry.instances = nil
		instantiator, ok := entry.feature.(FeatureInstantiator)
		if !ok {
			continue
		}

		apiType := instantiator.ServiceAPIType()
		if apiType == nil || apiType.Kind() != reflect.Interface {
			return fmt.Errorf("feature '%s' returned an invalid service API type", entry.name)
		}
		if owner, exists := claimedAPIs[apiType]; exists {
			return fmt.Errorf("features '%s' and '%s' both instantiate API %v", owner, entry.name, apiType)
		}
		claimedAPIs[apiType] = entry.name
	}

	for _, feature := range s.orderedFeatures {
		instantiator, supportsInstances := feature.feature.(FeatureInstantiator)
		matching := matchingDeclarations(declarations, instantiator, supportsInstances)
		if len(matching) == 0 {
			if err := s.initializeRegisteredFeature(ctx, feature, feature.feature, options, nil); err != nil {
				return err
			}
			continue
		}

		instances := []Feature{feature.feature}
		for i := 1; i < len(matching); i++ {
			instance := instantiator.NewInstance()
			if instance == nil {
				return fmt.Errorf("feature '%s' returned a nil instance", feature.name)
			}
			instances = append(instances, instance)
		}

		for i, instance := range instances {
			declaration := matching[i]
			if err := s.initializeRegisteredFeature(ctx, feature, instance, options, &declaration); err != nil {
				return err
			}
			s.declaredFeatures[declaration.FieldName] = instance
		}
	}

	return nil
}

func matchingDeclarations(
	declarations []FeatureDeclaration,
	instantiator FeatureInstantiator,
	supported bool,
) []FeatureDeclaration {
	if !supported {
		return nil
	}

	var result []FeatureDeclaration
	apiType := instantiator.ServiceAPIType()
	for _, declaration := range declarations {
		if declaration.APIType == apiType {
			result = append(result, declaration)
		}
	}

	return result
}

func (s *FeatureSet) initializeRegisteredFeature(
	ctx context.Context,
	registered *registeredFeature,
	feature Feature,
	options *InitializeOptions,
	declaration *FeatureDeclaration,
) error {
	allowOptions := &CanBeInitializedOptions{
		DeploymentEnv: options.Env.DeploymentEnv(),
		Definitions:   options.Definitions,
	}
	createOptions := &InitializeOptions{
		Logger:         options.Logger,
		Errors:         options.Errors,
		Definitions:    options.Definitions,
		Tags:           options.Tags,
		ServiceContext: options.ServiceContext,
		Dependencies:   s.getDependentFeatures(registered.dependencies),
		FeatureInputs:  options.FeatureInputs,
		Env:            options.Env,
		Declaration:    declaration,
	}

	feature.UpdateInfo(UpdateInfoEntry{Name: registered.name})
	if err := s.initializeFeature(ctx, feature, allowOptions, createOptions); err != nil {
		return err
	}
	registered.instances = append(registered.instances, feature)
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

func (s *FeatureSet) initializeFeature(
	ctx context.Context,
	feature Feature,
	allow *CanBeInitializedOptions,
	create *InitializeOptions,
) error {
	enabled := feature.CanBeInitialized(allow)
	feature.UpdateInfo(UpdateInfoEntry{
		Enabled: enabled,
		Logger:  create.Logger,
		Errors:  create.Errors,
	})

	if enabled {
		if err := feature.Initialize(ctx, create); err != nil {
			return err
		}
	}

	return nil
}

// Register registers an internal feature that will be initialized, if allowed,
// to be used by a service. The features will be initialized in the order they
// are registered.
//
// If a feature already exists with the same name, it will be replaced.
func (s *FeatureSet) Register(name string, feature Feature, dependencies ...string) {
	if feature == nil {
		return
	}

	// Gives the feature access to its name from this point on.
	feature.UpdateInfo(UpdateInfoEntry{Name: name})
	if entry, ok := s.features[name]; ok {
		entry.feature = feature
		entry.dependencies = dependencies
		return
	}

	f := &registeredFeature{
		dependencies: dependencies,
		feature:      feature,
		name:         name,
	}

	s.features[name] = f
	s.orderedFeatures = append(s.orderedFeatures, f)
}

// Feature retrieves the requested feature by its name if it has been registered.
func (s *FeatureSet) Feature(name string) (Feature, error) {
	feature, ok := s.features[name]
	if !ok {
		return nil, fmt.Errorf("could not find feature '%v'", name)
	}

	return feature.feature, nil
}

// FeatureForField returns the feature instance created for a tagged service
// field. The boolean is false for fields handled by singleton features.
func (s *FeatureSet) FeatureForField(fieldName string) (Feature, bool) {
	feature, ok := s.declaredFeatures[fieldName]
	return feature, ok
}

// Iterator returns a new FeatureSetIterator for iterating over the ordered
// features in the FeatureSet.
func (s *FeatureSet) Iterator() *FeatureSetIterator {
	var features []Feature
	for _, registered := range s.orderedFeatures {
		features = append(features, registered.lifecycleFeatures()...)
	}
	return &FeatureSetIterator{
		features: features,
		index:    0,
	}
}

// Count returns the total number of features registered in the FeatureSet.
func (s *FeatureSet) Count() int {
	return len(s.features)
}

// Append adds all registered features from the given FeatureSet into the current
// FeatureSet, maintaining their order.
func (s *FeatureSet) Append(features *FeatureSet) {
	if features != nil {
		for _, feature := range features.orderedFeatures {
			s.features[feature.name] = feature
			s.orderedFeatures = append(s.orderedFeatures, feature)
		}
	}
}

// StartAll iterates over all registered features and invokes their Start
// method if they implement FeatureController.
func (s *FeatureSet) StartAll(ctx context.Context, srv interface{}) error {
	for _, registered := range s.orderedFeatures {
		for _, feature := range registered.lifecycleFeatures() {
			if p, ok := feature.(FeatureController); ok {
				if err := p.Start(ctx, srv); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// CleanupAll iterates through all features and calls their Cleanup method
// if they implement FeatureController.
func (s *FeatureSet) CleanupAll(ctx context.Context) error {
	for _, registered := range s.orderedFeatures {
		for _, feature := range registered.lifecycleFeatures() {
			if p, ok := feature.(FeatureController); ok {
				if err := p.Cleanup(ctx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (f *registeredFeature) lifecycleFeatures() []Feature {
	if len(f.instances) == 0 {
		return []Feature{f.feature}
	}
	return f.instances
}

// Next retrieves the next Feature in the iteration. Returns the Feature and
// true if found.
func (i *FeatureSetIterator) Next() (Feature, bool) {
	if i.index < len(i.features) {
		feature := i.features[i.index]
		i.index++
		return feature, true
	}

	return nil, false
}

// Reset sets the iterator's index back to the beginning, allowing iteration to
// start over from the first feature.
func (i *FeatureSetIterator) Reset() {
	i.index = 0
}
