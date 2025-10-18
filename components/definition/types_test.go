package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedServiceTypes(t *testing.T) {
	t.Run("should have all supported services", func(t *testing.T) {
		types := SupportedServiceTypes()
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
