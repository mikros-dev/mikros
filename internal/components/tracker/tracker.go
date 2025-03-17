package tracker

import (
	"strings"

	trackerApi "github.com/mikros-dev/mikros/apis/tracker"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Tracker struct {
	tracker plugin.Feature
}

func New(features *plugin.FeatureSet) (*Tracker, error) {
	f, err := features.Feature(options.TrackerFeatureName)
	if err != nil && !strings.Contains(err.Error(), "could not find feature") {
		return nil, err
	}

	return &Tracker{
		tracker: f,
	}, nil
}

func (t *Tracker) Tracker() (trackerApi.Tracker, bool) {
	if t.tracker != nil {
		if api, ok := t.tracker.(plugin.FeatureInternalAPI); ok {
			return api.FrameworkAPI().(trackerApi.Tracker), true
		}
	}

	return nil, false
}
