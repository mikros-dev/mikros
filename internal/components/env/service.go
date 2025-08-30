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

	// Load service defined environment variables (through service.toml 'envs' key)
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

func (s *ServiceEnvs) DefinedEnv(name string) (string, bool) {
	v, ok := s.definedEnvs[name]
	return v, ok
}

func (s *ServiceEnvs) DeploymentEnv() definition.ServiceDeploy {
	return s.envs.DeploymentEnv
}

func (s *ServiceEnvs) TrackerHeaderName() string {
	return s.envs.TrackerHeaderName
}

func (s *ServiceEnvs) IsCICD() bool {
	return s.envs.IsCICD
}

func (s *ServiceEnvs) CoupledNamespace() string {
	return s.envs.CoupledNamespace
}

func (s *ServiceEnvs) CoupledPort() int32 {
	return s.envs.CoupledPort
}

func (s *ServiceEnvs) GrpcPort() int32 {
	return s.envs.GrpcPort
}

func (s *ServiceEnvs) HttpPort() int32 {
	return s.envs.HttpPort
}

func (s *ServiceEnvs) Get(key string) string {
	key = strings.TrimSuffix(key, stringEnvNotation)
	return s.definedEnvs[key]
}

func (s *ServiceEnvs) GetInt(name string) (int, error) {
	i, err := strconv.Atoi(s.Get(name))
	return i, err
}

func (s *ServiceEnvs) GetBool(name string) (bool, error) {
	b, err := strconv.ParseBool(s.Get(name))
	return b, err
}
