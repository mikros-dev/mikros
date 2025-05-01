package env

import (
	"fmt"
	"os"
	"testing"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

type envsExample struct {
	DeploymentEnv definition.ServiceDeploy `env:"SERVICE_DEPLOY,default_value=local"`
	AwsRegion     string                   `env:"AWS_REGION"`

	// CI/CD settings
	IsCICD bool `env:"CICD_TEST,default_value=false"`

	// Auth settings
	AuthPoolID Env[string] `env:"AUTH_POOL_ID"`
	Number     Env[int32]  `env:"NUMBER"`

	// Database settings
	DatabasePort int32 `env:"DATABASE_PORT,default_value=27017"`

	Empty string `env:",skip"`
}

func TestLoad(t *testing.T) {
	tests := []struct {
		Title     string
		Variables map[string]string
	}{
		{
			Title: "successfully load all fields",
			Variables: map[string]string{
				"SERVICE_DEPLOY": "dev",
				"AWS_REGION":     "us-east1",
				"CICD_TEST":      "true",
				"AUTH_POOL_ID":   "some-random-id",
				"NUMBER":         "42",
				"DATABASE_PORT":  "9999",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Title, func(t *testing.T) {
			for k, v := range test.Variables {
				_ = os.Setenv(k, v)
			}

			var e envsExample
			err := Load(service.FromString("example"), &e)

			fmt.Println(err)
			fmt.Println(e)

			for k := range test.Variables {
				_ = os.Unsetenv(k)
			}
		})
	}
}
