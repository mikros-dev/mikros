package definition

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefinitionsValidation(t *testing.T) {
	a := assert.New(t)
	tests := []struct {
		Title           string
		TomlDefinitions string
		Expected        []string
		DefsAssertion   func(object interface{}, msgAndArgs ...interface{}) bool
		ErrorAssertion  func(err error, msgAndArgs ...interface{}) bool
		CustomAssertion func(defs *Definitions)
	}{
		{
			Title: "should succeed with script type alone",
			TomlDefinitions: `
name = "example"
types = ["script"]
version = "v1.0.0"
language = "go"
product = "SDS"
`,
			ErrorAssertion: a.NoError,
		},
		{
			Title: "should not have script service type with other types",
			TomlDefinitions: `
name = "example"
types = ["grpc", "http", "script"]
version = "v1.0.0"
language = "go"
product = "SDS"
`,
			ErrorAssertion: a.Error,
			Expected: []string{
				"failed on the 'single_script' tag",
			},
		},
		{
			Title: "should not have duplicated service types",
			TomlDefinitions: `
name = "example"
types = ["grpc", "http", "http"]
version = "v1.0.0"
language = "go"
product = "SDS"
`,
			ErrorAssertion: a.Error,
			Expected: []string{
				"failed on the 'no_duplicated_service' tag",
			},
		},
		{
			Title: "should fail without setting service type",
			TomlDefinitions: `
name = "example"
version = "v1.0.0"
language = "go"
product = "SDS"
`,
			ErrorAssertion: a.Error,
			Expected: []string{
				"'Definitions.Types' Error:Field validation for 'Types' failed on the 'required' tag",
			},
		},
		{
			Title: "should fail with unsupported types",
			TomlDefinitions: `
name = "example"
types = ["grpc", "unsupported", "http"]
version = "v1.0.0"
language = "go"
product = "SDS"
`,
			ErrorAssertion: a.Error,
			Expected: []string{
				"'Definitions.Types[1]' Error:Field validation for 'Types[1]' failed on the 'service_type' tag",
			},
		},
		{
			Title: "should fail with unsupported service type",
			TomlDefinitions: `
name = "example"
types = ["unsupported"]
version = "v1.0.0"
language = "go"
product = "SDS"
`,
			ErrorAssertion: a.Error,
			Expected:       []string{"'Definitions.Types[0]' Error:Field validation for 'Types[0]' failed on the 'service_type' tag"},
		},
		{
			Title: "should fail with invalid input",
			TomlDefinitions: `
name = "service_test"
types = ["monolith"]
version = "5.1-alpha"
language = "java"
product = "UNKNOWN"
envs = [ "wrong_case" ]

[features.database]
kind = "sqlserver"
ttl = -1

[features.pubsub]
emitted_events = [ "UNSUPPORTED_EVENT1", "UNSUPPORTED_EVENT2", "" ]
`,
			Expected: []string{
				"'Definitions.Types[0]' Error:Field validation for 'Types[0]' failed on the 'service_type' tag\n",
				"'Definitions.Version' Error:Field validation for 'Version' failed on the 'version' tag",
				"'Definitions.Language' Error:Field validation for 'Language' failed on the 'oneof' tag",
				"'Definitions.Product' Error:Field validation for 'Product' failed on the 'oneof' tag",
				"'Definitions.Envs[0]' Error:Field validation for 'Envs[0]' failed on the 'uppercase' tag",
			},
			ErrorAssertion: a.Error,
			DefsAssertion:  a.NotNil,
		},
		{
			Title: "should fail without version",
			TomlDefinitions: `
name = "service_test"
types = ["grpc"]
language = "go"
product = "SDS"
envs = [ "REGION" ]
emitted_events = [ "VEHICLE_CREATED" ]
`,
			Expected: []string{
				"'Definitions.Version' Error:Field validation for 'Version' failed on the 'required' tag",
			},
			DefsAssertion:  a.NotNil,
			ErrorAssertion: a.Error,
		},
		{
			Title: "succeed with service custom settings",
			TomlDefinitions: `
name = "service_test"
types = ["http"]
version = "v0.1.0"
language = "go"
product = "SDS"
envs = [ "REGION" ]

[service]
value = 42
password = "Hello"
`,
			DefsAssertion:  a.NotNil,
			ErrorAssertion: a.NoError,
			Expected:       []string{"teste"},
			CustomAssertion: func(defs *Definitions) {
				a.NotNil(defs.Service)
			},
		},
		{
			Title: "succeed with custom clients settings",
			TomlDefinitions: `
name = "service_test"
types = ["grpc"]
version = "v0.1.0"
language = "go"
product = "SDS"
envs = [ "REGION" ]

[clients.contract]
host = "localhost"
port = 9192

`,
			DefsAssertion:  a.NotNil,
			ErrorAssertion: a.NoError,
			Expected:       []string{"teste"},
			CustomAssertion: func(defs *Definitions) {
				a.Equal(1, len(defs.Clients))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Title, func(t *testing.T) {
			tmpFile, _ := os.CreateTemp(os.TempDir(), "pre-*.toml")
			defer func() { _ = os.Remove(tmpFile.Name()) }()
			_, _ = tmpFile.Write([]byte(test.TomlDefinitions))
			_ = tmpFile.Close()

			defs, _ := ParseFromFile(tmpFile.Name())
			err := defs.Validate()

			if test.DefsAssertion != nil {
				test.DefsAssertion(defs)
			}

			if test.ErrorAssertion != nil {
				test.ErrorAssertion(err)
			}

			if err != nil {
				for _, expected := range test.Expected {
					a.Contains(err.Error(), expected)
				}
			}

			if defs != nil && test.CustomAssertion != nil {
				test.CustomAssertion(defs)
			}
		})
	}
}
