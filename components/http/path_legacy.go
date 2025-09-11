//go:build !go1.22

package http

import (
	"net/http"
)

// StdPathGetter is a no-op on Go < 1.22; callers can override via BindOptions.
func StdPathGetter(_ *http.Request, _ string) (string, bool) {
	return "", false
}
