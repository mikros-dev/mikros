package http

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/creasty/defaults"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/options"
)

type Definitions struct {
	CORSStrict     bool          `toml:"cors_strict" json:"cors_strict" default:"true"`
	DisableAuth    bool          `toml:"disable_auth,omitempty" json:"disable_auth" default:"false"`
	BasePath       string        `toml:"base_path" json:"base_path"`
	ReadTimeout    time.Duration `toml:"read_timeout" json:"read_timeout" default:"15s"`
	WriteTimeout   time.Duration `toml:"write_timeout" json:"write_timeout" default:"15s"`
	IdleTimeout    time.Duration `toml:"idle_timeout" json:"idle_timeout" default:"60s"`
	MaxHeaderBytes int           `toml:"max_header_bytes" json:"max_header_bytes" default:"1048576"`
}

func newDefinitions(definitions *definition.Definitions, opt *options.HttpServiceOptions) *Definitions {
	out := &Definitions{}
	_ = defaults.Set(out)

	// Apply programmatic options
	if opt != nil {
		bp := normalizeBasePath(opt.BasePath)
		if bp != "" {
			out.BasePath = bp
		}

		mergeNonZero(out, opt)
	}

	// Apply file definitions
	if currentDefs, ok := definitions.LoadService(definition.ServiceType_HTTP); ok {
		if b, err := json.Marshal(currentDefs); err == nil {
			var defs Definitions
			if json.Unmarshal(b, &defs) == nil {
				// File version of the following settings always wins
				out.DisableAuth = defs.DisableAuth
				out.CORSStrict = defs.CORSStrict

				// Only use the file version if it's not empty'
				if defs.BasePath != "" {
					out.BasePath = normalizeBasePath(defs.BasePath)
				}

				mergeNonZero(out, &defs)
			}
		}
	}

	return out
}

// normalizeBasePath ensures a leading "/" and trims trailing "/".
// "" or "/" will normalize to "" (mounted at root).
func normalizeBasePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimRight(p, "/")
}

// mergeNonZero copies non-zero fields from src into dst by field name.
func mergeNonZero(dst, src interface{}) {
	var (
		dv = reflect.ValueOf(dst).Elem()
		sv = reflect.ValueOf(src).Elem()
		st = sv.Type()
	)

	for i := 0; i < sv.NumField(); i++ {
		var (
			sf     = sv.Field(i)
			sfType = st.Field(i)
			name   = sfType.Name
		)

		// Skip unexported or missing counterpart on dst
		if sfType.PkgPath != "" {
			continue
		}

		df := dv.FieldByName(name)
		if !df.IsValid() || !df.CanSet() {
			continue
		}

		switch df.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			if sf.Int() > 0 {
				df.SetInt(sf.Int())
			}
		default:
			// Ignore other types by now
		}
	}
}
