package plugin

// RuntimeSet represents a collection of runtimes, providing mechanisms to
// register and manage them.
type RuntimeSet struct {
	runtimes map[string]Runtime
}

// NewRuntimeSet creates and returns a new instance of RuntimeSet with an
// initialized runtimes map.
func NewRuntimeSet() *RuntimeSet {
	return &RuntimeSet{
		runtimes: make(map[string]Runtime),
	}
}

// Register adds a runtime to the RuntimeSet if it has a valid name and is not
// yet registered.
func (r *RuntimeSet) Register(runtime Runtime) {
	if name := runtime.Name(); name != "" {
		if _, ok := r.runtimes[name]; !ok {
			r.runtimes[name] = runtime
		}
	}
}

// Runtimes returns a copy of the registered runtimes within the RuntimeSet.
func (r *RuntimeSet) Runtimes() map[string]Runtime {
	svc := make(map[string]Runtime)
	for k, v := range r.runtimes {
		svc[k] = v
	}

	return svc
}

// Append adds runtimes from another RuntimeSet to the current one if they are
// not already present.
func (r *RuntimeSet) Append(runtimes *RuntimeSet) {
	for k, v := range runtimes.Runtimes() {
		if _, ok := r.runtimes[k]; !ok {
			r.runtimes[k] = v
		}
	}
}
