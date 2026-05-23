package integrations

import (
	"github.com/mikros-dev/mikros/components/plugin"
)

// Integrations creates the integration set for the mikros framework.
func Integrations() *plugin.IntegrationSet {
	return plugin.NewIntegrationSet()
}
