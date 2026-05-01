package tracker

import (
	"strings"

	"github.com/mikros-dev/mikros/apis/integrations"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Tracker is a type that encapsulates the tracker plugin integration, allowing
// integration with external or internal frameworks.
type Tracker struct {
	tracker plugin.Integration
}

// New creates a new Tracker instance using the specified IntegrationSet to search
// for the tracker integration.
func New(integrations *plugin.IntegrationSet) (*Tracker, error) {
	i, err := integrations.Integration(options.TrackerIntegrationName)
	if err != nil && !strings.Contains(err.Error(), "could not find integration") {
		return nil, err
	}

	return &Tracker{
		tracker: i,
	}, nil
}

// Tracker retrieves the tracker implementation if available.
func (t *Tracker) Tracker() (integrations.Tracker, bool) {
	if t.tracker != nil {
		v, ok := t.tracker.API().(integrations.Tracker)
		return v, ok
	}

	return nil, false
}
