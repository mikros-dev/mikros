package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedRuntimeTypes(t *testing.T) {
	t.Run("should have all supported runtimes", func(t *testing.T) {
		types := SupportedRuntimeTypes()
		a := assert.New(t)
		a.Equal(5, len(types))
	})
}

func TestSupportedLanguages(t *testing.T) {
	t.Run("should have all supported languages", func(t *testing.T) {
		languages := SupportedLanguages()
		a := assert.New(t)
		a.Equal(2, len(languages))
	})
}
