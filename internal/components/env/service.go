package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/env"
)

const (
	stringEnvNotation = "@env"
)

// ServiceEnvs is the object that will allow all internal (and external) mikros
// features to access the environment variables loaded.
type ServiceEnvs struct {
	envs *GlobalEnvs

	// definedEnvs holds all variables pointed directly into the 'service.toml'
	// file.
	definedEnvs map[string]string `env:",skip"`
}

// NewServiceEnvs loads the framework main environment variables through the env
// feature plugin.
func NewServiceEnvs(defs *definition.Definitions) (*ServiceEnvs, error) {
	var envs GlobalEnvs
	if err := env.Load(defs.ServiceName(), &envs); err != nil {
		return nil, err
	}

	envs.postLoad()

	// Load service-defined environment variables (through service.toml 'envs' key)
	definedEnvs, err := loadDefinedEnvVars(defs)
	if err != nil {
		return nil, err
	}

	return &ServiceEnvs{
		envs:        &envs,
		definedEnvs: definedEnvs,
	}, nil
}

// loadDefinedEnvVars loads envs defined in the 'service.toml' file as mandatory
// values. They must be available when the service starts.
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

// DefinedEnv retrieves the value of a specific environment variable by name
// from the defined envs in the service.toml file.
func (s *ServiceEnvs) DefinedEnv(name string) (string, bool) {
	v, ok := s.definedEnvs[name]
	return v, ok
}

// DeploymentEnv retrieves the deployment environment of the service.
func (s *ServiceEnvs) DeploymentEnv() definition.ServiceDeploy {
	return s.envs.DeploymentEnv
}

// TrackerHeaderName retrieves the tracker header name from the environment
// configuration.
func (s *ServiceEnvs) TrackerHeaderName() string {
	return s.envs.TrackerHeaderName
}

// IsCICD checks if the current environment is running in a CI/CD pipeline
// based on the environment configuration.
func (s *ServiceEnvs) IsCICD() bool {
	return s.envs.IsCICD
}

// CoupledNamespace retrieves the namespace configuration for coupled services
// from the environment.
func (s *ServiceEnvs) CoupledNamespace() string {
	return s.envs.CoupledNamespace
}

// CoupledPort retrieves the port configuration for coupled services from the
// environment variables.
func (s *ServiceEnvs) CoupledPort() int32 {
	return s.envs.CoupledPort
}

// GrpcPort retrieves the gRPC port configuration defined in the environment
// variables.
func (s *ServiceEnvs) GrpcPort() int32 {
	return s.envs.GrpcPort
}

// HTTPPort retrieves the HTTP port configuration value from the environment
// variables.
func (s *ServiceEnvs) HTTPPort() int32 {
	return s.envs.HTTPPort
}

// Get retrieves the value of a specified key from the defined environment
// variables.
func (s *ServiceEnvs) Get(key string) string {
	key = strings.TrimSuffix(key, stringEnvNotation)
	return s.definedEnvs[key]
}

// GetInt retrieves the integer value of the specified environment variable.
func (s *ServiceEnvs) GetInt(name string) (int, error) {
	i, err := strconv.Atoi(s.Get(name))
	return i, err
}

// GetBool retrieves the boolean value of the specified environment variable.
func (s *ServiceEnvs) GetBool(name string) (bool, error) {
	b, err := strconv.ParseBool(s.Get(name))
	return b, err
}
