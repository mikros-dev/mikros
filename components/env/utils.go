package env

import (
	"os"
	"strings"
)

const (
	stringEnvNotation = "@env"
)

// HasEnvNotation checks if a string has the mikros framework env notation
// indicating that it should be loaded from environment variables.
func HasEnvNotation(s string) bool {
	return strings.HasSuffix(s, stringEnvNotation)
}

// GetEnv is a helper function that retrieves a value from an environment
// variable independently if is has the env notation or not.
func GetEnv(s string) string {
	return os.Getenv(strings.TrimSuffix(s, stringEnvNotation))
}
