package validations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureValuesAreInitialized(t *testing.T) {
	a := assert.New(t)

	t.Run("nil value", func(t *testing.T) {
		err := EnsureValuesAreInitialized(nil)
		a.NotNil(err)
	})

	t.Run("empty values", func(t *testing.T) {
		type Server struct {
			Host string
			Port int
		}

		s := Server{}
		err := EnsureValuesAreInitialized(s)
		a.NotNil(err)
	})

	t.Run("attempt to validate other types", func(t *testing.T) {
		var err error

		i := 1
		err = EnsureValuesAreInitialized(i)
		a.NotNil(err)
		err = EnsureValuesAreInitialized(&i)
		a.NotNil(err)

		s := "Hello World!"
		err = EnsureValuesAreInitialized(s)
		a.NotNil(err)
		err = EnsureValuesAreInitialized(&s)
		a.NotNil(err)

		m := map[string]interface{}{
			"key":   "answer",
			"value": 42,
		}
		err = EnsureValuesAreInitialized(m)
		a.NotNil(err)
		err = EnsureValuesAreInitialized(&m)
		a.NotNil(err)
	})

	t.Run("some with empty values", func(t *testing.T) {
		type Server struct {
			Host           string
			Port           int
			MaxConnections uint
			URI            string
		}

		var (
			s   Server
			err error
		)

		s = Server{
			Host:           "www.example.com",
			MaxConnections: 10,
			URI:            "/index.html",
		}

		err = EnsureValuesAreInitialized(s)
		a.NotNil(err)
		a.True(strings.Contains(err.Error(), "could not initiate struct Server, value from field Port is missing"))

		s = Server{
			Host: "www.example.com",
			Port: 443,
			URI:  "/index.html",
		}

		err = EnsureValuesAreInitialized(s)
		a.NotNil(err)
		a.True(strings.Contains(err.Error(), "could not initiate struct Server, value from field MaxConnections is missing"))
	})

	t.Run("with empty pointer values", func(t *testing.T) {
		type Server struct {
			Host string
			Port int
			URI  *string
		}

		var (
			s   Server
			err error
		)

		s = Server{
			Host: "www.example.com",
			Port: 80,
		}

		err = EnsureValuesAreInitialized(s)
		a.NotNil(err)
		a.True(strings.Contains(err.Error(), "could not initiate struct Server, value from field URI is missing"))

	})

	t.Run("with all initialized", func(t *testing.T) {
		type Server struct {
			Host string
			Port int
		}

		s := Server{
			Host: "www.example.com",
			Port: 443,
		}

		err := EnsureValuesAreInitialized(s)
		a.Nil(err)
	})

	t.Run("with all initialized as pointer", func(t *testing.T) {
		type Server struct {
			Host string
			Port int
			URI  *string
		}

		uri := "/index.html"
		s := &Server{
			Host: "www.example.com",
			Port: 443,
			URI:  &uri,
		}

		err := EnsureValuesAreInitialized(s)
		a.Nil(err)
	})
}
