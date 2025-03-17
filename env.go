package mikros

import (
	"fmt"
	"os"
	"strings"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/env"
)

const (
	stringEnvNotation = "@env"
)

// Env is the main framework environment structure. It holds only variables
// common for the whole project.
//
// It is also the mechanism to hold all environment variables declared directly
// inside the 'service.toml' file.
type Env struct {
	DeploymentEnv     definition.ServiceDeploy `env:"MIKROS_SERVICE_DEPLOY,default_value=local"`
	TrackerHeaderName string                   `env:"MIKROS_TRACKER_HEADER_NAME,default_value=X-Request-ID"`

	// CI/CD settings
	IsCICD bool `env:"MIKROS_CICD_TEST,default_value=false"`

	// Coupled clients
	CoupledNamespace string `env:"MIKROS_COUPLED_NAMESPACE"`
	CoupledPort      int32  `env:"MIKROS_COUPLED_PORT,default_value=7070"`

	// Default connection ports
	GrpcPort int32 `env:"MIKROS_GRPC_PORT,default_value=7070"`
	HttpPort int32 `env:"MIKROS_HTTP_PORT,default_value=8080"`

	// definedEnvs holds all variables pointed directly into the 'service.toml'
	// file.
	definedEnvs map[string]string `env:",skip"`
}

func newEnv(defs *definition.Definitions) (*Env, error) {
	var envs Env
	if err := env.Load(defs.ServiceName(), &envs); err != nil {
		return nil, err
	}

	envs.autoAdjust()

	// Load service defined environment variables (through service.toml 'envs' key)
	definedEnvs, err := loadDefinedEnvVars(defs)
	if err != nil {
		return nil, err
	}
	envs.definedEnvs = definedEnvs

	return &envs, nil
}

// loadDefinedEnvVars loads envs defined in the 'service.toml' file as mandatory
// values, í.e., they must be available when the service starts.
func loadDefinedEnvVars(defs *definition.Definitions) (map[string]string, error) {
	var (
		envs = make(map[string]string)
	)

	for _, e := range defs.Envs {
		v, err := mustGetEnv(e)
		if err != nil {
			return nil, err
		}

		envs[e] = v
	}

	return envs, nil
}

// mustGetEnv retrieves a value from an environment variable and aborts
// if it is not set.
func mustGetEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("environment variable '%v' must be set", name)
	}

	return value, nil
}

// autoAdjust verifies if the local environment has any modification that
// needs to be reflected in the structure members.
func (e *Env) autoAdjust() {
	// Checks our real deployment environment
	if e.isRunningTest() {
		e.DeploymentEnv = definition.ServiceDeploy_Test
	}
}

// isRunningTest returns if the current session is being executed in test mode.
func (e *Env) isRunningTest() bool {
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") || strings.Contains(arg, "-test") {
			return true
		}
	}

	return false
}

func (e *Env) DefinedEnv(name string) (string, bool) {
	v, ok := e.definedEnvs[name]
	return v, ok
}

func (e *Env) ToMapEnv() *MapEnv {
	return &MapEnv{
		env: e,
	}
}

// MapEnv is an Env subtype that can be passed to the plugin package options.
type MapEnv struct {
	env *Env
}

func (m *MapEnv) DeploymentEnv() definition.ServiceDeploy {
	return m.env.DeploymentEnv
}

func (m *MapEnv) TrackerHeaderName() string {
	return m.env.TrackerHeaderName
}

func (m *MapEnv) IsCICD() bool {
	return m.env.IsCICD
}

func (m *MapEnv) CoupledNamespace() string {
	return m.env.CoupledNamespace
}

func (m *MapEnv) CoupledPort() int32 {
	return m.env.CoupledPort
}

func (m *MapEnv) GrpcPort() int32 {
	return m.env.GrpcPort
}

func (m *MapEnv) HttpPort() int32 {
	return m.env.HttpPort
}

func (m *MapEnv) Get(key string) interface{} {
	key = strings.TrimSuffix(key, stringEnvNotation)
	return m.env.definedEnvs[key]
}

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
