// Package env loads environment variables into a struct using field tags.
//
// Overview
//
//   - Tag syntax:        `env:"NAME[,default_value=VAL][,required]"`
//   - Precedence:        SERVICE<sep>NAME → NAME (service-scoped overrides global)
//   - Default separator: "__" (portable); can be changed via Options
//   - Pointer fields:    rejected when tagged (use value types or Env[T])
//   - Missing values:    if `required` and not found (and no default) → error
//     otherwise leave zero value (or zero Env[T] capturing VarName)
//   - Supported types:   string, bool, int/int32/int64, uint/uint32/uint64,
//     float32/float64, time.Duration,
//     and custom types implementing encoding.TextUnmarshaler.
//
// # Service-scoped precedence
//
// Load considers a service-specific variable first, then the global name.
// For service "file" and default separator "__":
//
//	file__DB_HOST  // checked first
//	DB_HOST        // fallback
//
// # Custom separator
//
// The default separator is "__". To change it, pass an Options value:
//
//	_ = env.Load(service.FromString("file"), &cfg, env.Options{Separator: "::"})
//
// Env[T] wrappers
//
// Env[T] captures both the parsed value and the concrete environment variable
// name used (via VarName). Supported instantiations are Env[string] and Env[int32].
//
// When a variable is not found and no default is provided, scalar fields keep
// their zero value. For Env[T], a zero-valued wrapper is assigned and VarName
// records the resolved key.
//
// # Pointers are not supported
//
// Tagged pointer fields (e.g., *int, *MyType) are rejected to avoid nil vs.
// zero-value ambiguity and implicit allocation. Use a value field or wrap in
// Env[T] if presence/source tracking is needed.
//
// Examples
//
//	type Config struct {
//	    Region      string          `env:"AWS_REGION,required"`
//	    Port        int32           `env:"DB_PORT,default_value=5432"`
//	    TTL         time.Duration   `env:"CACHE_TTL,default_value=30s"`
//	    PoolID      env.Env[string] `env:"AUTH_POOL_ID"`
//	}
//
//	var cfg Config
//	if err := env.Load(service.FromString("file"), &cfg); err != nil {
//	    // handle error (missing required, parse failure, unsupported type, etc.)
//	}
package env
