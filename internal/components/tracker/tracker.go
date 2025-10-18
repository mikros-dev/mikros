package tracker

import (
	"strings"

	"github.com/mikros-dev/mikros/apis/behavior"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Tracker is a type that encapsulates the tracker plugin feature, allowing
// integration with external or internal frameworks.
type Tracker struct {
	tracker plugin.Feature
}

// New creates a new Tracker instance using the specified FeatureSet to search
// for the tracker feature.
func New(features *plugin.FeatureSet) (*Tracker, error) {
	f, err := features.Feature(options.TrackerFeatureName)
	if err != nil && !strings.Contains(err.Error(), "could not find feature") {
		return nil, err
	}

	return &Tracker{
		tracker: f,
	}, nil
}

// Tracker retrieves the tracker implementation if available.
func (t *Tracker) Tracker() (behavior.Tracker, bool) {
	if t.tracker != nil {
		if api, ok := t.tracker.(plugin.FeatureInternalAPI); ok {
			return api.FrameworkAPI().(behavior.Tracker), true
		}
	}

	return nil, false
}
