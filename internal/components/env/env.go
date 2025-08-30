package env

import (
	"os"
	"strings"

	"github.com/mikros-dev/mikros/components/definition"
)

// GlobalEnvs is the main framework structure that holds environment variables. Main
// variables are declared as structure member and are loaded directly, using
// struct tags.
type GlobalEnvs struct {
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
}

// postLoad is where any internal change must happen, according to the current
// environment values previously loaded.
func (e *GlobalEnvs) postLoad() {
	// Checks our real deployment environment
	if e.isRunningTest() {
		e.DeploymentEnv = definition.ServiceDeploy_Test
	}
}

// isRunningTest returns if the current session is being executed in test mode.
func (e *GlobalEnvs) isRunningTest() bool {
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") || strings.Contains(arg, "-test") {
			return true
		}
	}

	return false
}
