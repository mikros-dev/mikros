package plugin

import (
	"context"
	"fmt"
)

// IntegrationSet gathers all integrations that a service can use during its
// execution.
type IntegrationSet struct {
	integrations        map[string]*registeredIntegration
	orderedIntegrations []*registeredIntegration
}

type registeredIntegration struct {
	name        string
	integration Integration
}

// NewIntegrationSet creates a new IntegrationSet.
func NewIntegrationSet() *IntegrationSet {
	return &IntegrationSet{
		integrations: make(map[string]*registeredIntegration),
	}
}

// InitializeAll initializes all previously registered integrations in the
// order they were registered.
func (s *IntegrationSet) InitializeAll(ctx context.Context, options *InitializeOptions) error {
	for _, integration := range s.orderedIntegrations {
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
			FeatureInputs:  options.FeatureInputs,
			Env:            options.Env,
		}

		if err := s.initializeIntegration(ctx, integration.integration, allowOptions, createOptions); err != nil {
			return err
		}
	}

	return nil
}

func (s *IntegrationSet) initializeIntegration(
	ctx context.Context,
	integration Integration,
	allow *CanBeInitializedOptions,
	create *InitializeOptions,
) error {
	enabled := integration.CanBeInitialized(allow)
	integration.UpdateInfo(UpdateInfoEntry{
		Enabled: enabled,
		Logger:  create.Logger,
		Errors:  create.Errors,
	})

	if enabled {
		if err := integration.Initialize(ctx, create); err != nil {
			return err
		}
	}

	return nil
}

// Register registers an internal integration that will be initialized, if
// allowed, to be used by a service. The integrations will be initialized in
// the order they are registered.
//
// If an integration already exists with the same name, it will be replaced.
func (s *IntegrationSet) Register(name string, integration Integration) {
	if integration == nil {
		return
	}

	integration.UpdateInfo(UpdateInfoEntry{Name: name})
	if entry, ok := s.integrations[name]; ok {
		entry.integration = integration
		return
	}

	entry := &registeredIntegration{
		integration: integration,
		name:        name,
	}

	s.integrations[name] = entry
	s.orderedIntegrations = append(s.orderedIntegrations, entry)
}

// Integration retrieves the requested integration by its name if it has been
// registered.
func (s *IntegrationSet) Integration(name string) (Integration, error) {
	integration, ok := s.integrations[name]
	if !ok {
		return nil, fmt.Errorf("could not find integration '%v'", name)
	}

	return integration.integration, nil
}

// Count returns the total number of integrations registered in the
// IntegrationSet.
func (s *IntegrationSet) Count() int {
	return len(s.integrations)
}

// Append adds all registered integrations from the given IntegrationSet into
// the current IntegrationSet, maintaining their order.
func (s *IntegrationSet) Append(integrations *IntegrationSet) {
	if integrations != nil {
		for _, integration := range integrations.orderedIntegrations {
			s.integrations[integration.name] = integration
			s.orderedIntegrations = append(s.orderedIntegrations, integration)
		}
	}
}

// StartAll iterates over all registered integrations and invokes their Start
// method if they implement IntegrationController.
func (s *IntegrationSet) StartAll(ctx context.Context, srv interface{}) error {
	for _, integration := range s.integrations {
		if p, ok := integration.integration.(IntegrationController); ok {
			if err := p.Start(ctx, srv); err != nil {
				return err
			}
		}
	}

	return nil
}

// CleanupAll iterates through all integrations and calls their Cleanup method
// if they implement IntegrationController.
func (s *IntegrationSet) CleanupAll(ctx context.Context) error {
	for _, integration := range s.integrations {
		if p, ok := integration.integration.(IntegrationController); ok {
			if err := p.Cleanup(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
