package env

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

type baseConfig struct {
	DeployEnv   definition.ServiceDeploy `env:"DEPLOY_ENV"`
	Region      string                   `env:"AWS_REGION"`
	CI          bool                     `env:"CI,default_value=false"`
	Port        int                      `env:"PORT,default_value=8080"`
	Port32      int32                    `env:"DB_PORT,default_value=27017"`
	Port64      int64                    `env:"DB_PORT_64,default_value=3306"`
	RequiredKey string                   `env:"REQUIRED_KEY,required"`
	Pool        Env[string]              `env:"AUTH_POOL_ID"`
	N           Env[int32]               `env:"NUMBER"`
	TTL         time.Duration            `env:"CACHE_TTL,default_value=30s"`
	Speed       float32                  `env:"SPEED"`
	ExpireTime  float64                  `env:"EXPIRE_TIME"`
	Cost        uint                     `env:"COST"`
	Cost32      uint32                   `env:"COST32"`
	Cost64      uint64                   `env:"COST64"`
	Ignored     string
	private     string `env:"PRIVATE"`
}

func TestLoad(t *testing.T) {
	var (
		svc = service.FromString("example")
		a   = assert.New(t)
	)

	t.Run("successfully loads with defaults", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("CI", "true")
		t.Setenv("AUTH_POOL_ID", "pool-xyz")
		t.Setenv("NUMBER", "42")
		t.Setenv("REQUIRED_KEY", "present")
		t.Setenv("SPEED", "42.5")
		t.Setenv("EXPIRE_TIME", "10")
		t.Setenv("COST", "100")
		t.Setenv("COST32", "100")
		t.Setenv("COST64", "100")
		t.Setenv("DEPLOY_ENV", "test")
		t.Setenv("DB_PORT_64", "9981")
		t.Setenv("PORT", "8081")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.DeployEnv, definition.ServiceDeployTest)
		a.Equal(cfg.Region, "us-east-1")
		a.Equal(cfg.CI, true)
		a.Equal(cfg.Port, 8081)
		a.Equal(cfg.Port32, int32(27017))
		a.Equal(cfg.Port64, int64(9981))
		a.Equal(cfg.RequiredKey, "present")
		a.Equal(cfg.Pool.Value(), "pool-xyz")
		a.Equal(cfg.N.Value(), int32(42))
		a.Equal(cfg.TTL, time.Second*30)
		a.Equal(cfg.Speed, float32(42.5))
		a.Equal(cfg.ExpireTime, float64(10))
		a.Equal(cfg.Cost, uint(100))
		a.Equal(cfg.Cost32, uint32(100))
		a.Equal(cfg.Cost64, uint64(100))
		a.Equal(cfg.Ignored, "")
		a.Equal(cfg.private, "")
	})

	t.Run("required missing errors", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.NotNil(err)
		a.ErrorContains(err, "REQUIRED_KEY")
	})

	t.Run("service precedence with default separator", func(t *testing.T) {
		svc := service.FromString("file")
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("file__AWS_REGION", "eu-west-1")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.Region, "eu-west-1")
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("custom separator", func(t *testing.T) {
		svc := service.FromString("app")
		t.Setenv("app::AWS_REGION", "ap-south-1")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg, Options{Separator: "::"})

		a.Nil(err)
		a.Equal(cfg.Region, "ap-south-1")
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("bool parsing variants", func(t *testing.T) {
		t.Setenv("CI", "1")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.CI, true)
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("Env wrapper captures global var name", func(t *testing.T) {
		t.Setenv("AUTH_POOL_ID", "global-pool")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.Pool.Value(), "global-pool")
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("Env wrapper captures service-scoped var name", func(t *testing.T) {
		svc := service.FromString("svc")
		t.Setenv("svc__AUTH_POOL_ID", "svc-pool")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.Pool.Value(), "svc-pool")
		a.Equal(cfg.Pool.VarName(), "svc__AUTH_POOL_ID")
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("duration via time.Duration", func(t *testing.T) {
		t.Setenv("CACHE_TTL", "90s")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.TTL, time.Second*90)
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("skip tag ignores field even if set", func(t *testing.T) {
		t.Setenv("IGNORED", "should-not-be-set")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		err := Load(svc, &cfg)

		a.Nil(err)
		a.Equal(cfg.Ignored, "")
		a.Equal(cfg.RequiredKey, "present")
	})

	t.Run("invalid tags cause errors", func(t *testing.T) {
		type bad struct {
			Bad1 string `env:""`
		}

		var cfg bad
		err := Load(svc, &cfg)

		a.NotNil(err)
		a.ErrorContains(err, "'env' tag cannot be empty")
	})

	t.Run("target validation errors", func(t *testing.T) {
		var notPtr baseConfig
		err := Load(svc, notPtr)
		a.NotNil(err)

		var nilPtr *baseConfig
		err = Load(svc, nilPtr)
		a.NotNil(err)

		type notStruct int
		var x notStruct
		err = Load(svc, &x)
		a.NotNil(err)
	})

	t.Run("target with tagged pointer type", func(t *testing.T) {
		var example struct {
			Ex1 string  `env:"ex1"`
			Ex2 string  `env:"ex2"`
			Ex3 *string `env:"ex3"`
		}

		err := Load(svc, &example)
		a.Error(err)
		a.ErrorContains(err, "env: pointer-typed fields are not supported; use value type or Env[T]")
	})

	t.Run("target with convertible types", func(t *testing.T) {
		type Port int32
		type Label string

		type convConfig struct {
			DBPort Port  `env:"DB_PORT,default_value=5432"`
			Name   Label `env:"APP_NAME,default_value=hello"`
		}

		var cfg convConfig
		err := Load(service.FromString("svc"), &cfg)

		a.Nil(err)
		a.Equal(cfg.DBPort, Port(5432))
		a.Equal(cfg.Name, Label("hello"))
	})

	t.Run("default value with quotes", func(t *testing.T) {
		var example struct {
			Value string `env:"VALUE,default_value=\"some value and more\""`
		}

		err := Load(svc, &example)
		a.Nil(err)
		a.Equal(example.Value, "some value and more")
	})

	t.Run("should fail with empty default value", func(t *testing.T) {
		var example struct {
			Value string `env:"VALUE,default_value"`
		}

		err := Load(svc, &example)
		a.NotNil(err)
		a.ErrorContains(err, "default_value requires a value")
	})
}
