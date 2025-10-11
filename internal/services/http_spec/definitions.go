//revive:disable:var-naming
package http_spec

import (
	"encoding/json"

	"github.com/creasty/defaults"

	"github.com/mikros-dev/mikros/components/definition"
)

// Definitions contains configuration settings for the HTTP service behavior.
//
//revive:disable:line-length-limit
type Definitions struct {
	DisableAuth          bool `toml:"disable_auth,omitempty" default:"false" json:"disable_auth"`
	DisablePanicRecovery bool `toml:"disable_panic_recovery,omitempty" default:"false" json:"disable_panic_recovery"`
	MaxRequestBodySize   int  `toml:"max_request_body_size,omitempty" default:"4" json:"max_request_body_size"` // in megabytes
}

//revive:enable:line-length-limit

func newDefinitions(definitions *definition.Definitions) *Definitions {
	if currentDefs, ok := definitions.LoadService(definition.ServiceTypeHTTPSpec); ok {
		if b, err := json.Marshal(currentDefs); err == nil {
			var serviceDefs Definitions
			if err := json.Unmarshal(b, &serviceDefs); err == nil {
				return &serviceDefs
			}
		}
	}

	// Use the default values
	defs := &Definitions{}
	_ = defaults.Set(defs)

	return defs
}
