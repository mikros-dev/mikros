//go:build go1.22

package http

import (
	"net/http"
)

// StdPathGetter uses net/http's PathValue (Go 1.22+) to parse endpoint path
// parameters.
func StdPathGetter(r *http.Request, name string) (string, bool) {
	if v := r.PathValue(name); v != "" {
		return v, true
	}

	return "", false
}
