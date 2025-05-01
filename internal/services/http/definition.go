package http

import (
	"encoding/json"

	"github.com/creasty/defaults"

	"github.com/mikros-dev/mikros/components/definition"
)

type Definitions struct {
	DisableAuth          bool `toml:"disable_auth,omitempty" default:"false" json:"disable_auth"`
	DisablePanicRecovery bool `toml:"disable_panic_recovery,omitempty" default:"false" json:"disable_panic_recovery"`
}

func newDefinitions(definitions *definition.Definitions) *Definitions {
	if currentDefs, ok := definitions.LoadService(definition.ServiceType_HTTP); ok {
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
