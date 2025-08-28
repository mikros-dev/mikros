package env

import (
	"testing"
	"time"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

type baseConfig struct {
	DeployEnv   definition.ServiceDeploy `env:"DEPLOY_ENV"`
	Region      string                   `env:"AWS_REGION"`
	CI          bool                     `env:"CI,default_value=false"`
	Port        int32                    `env:"DB_PORT,default_value=27017"`
	RequiredKey string                   `env:"REQUIRED_KEY,required"`
	Pool        Env[string]              `env:"AUTH_POOL_ID"`
	N           Env[int32]               `env:"NUMBER"`
	TTL         time.Duration            `env:"CACHE_TTL,default_value=30s"`
	Speed       float32                  `env:"SPEED"`
	ExpireTime  float64                  `env:"EXPIRE_TIME"`
	Cost        uint                     `env:"COST"`
	Ignored     string
	private     string `env:"PRIVATE"`
}

func TestLoad(t *testing.T) {
	svc := service.FromString("example")

	t.Run("successfully loads with defaults", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("CI", "true")
		t.Setenv("AUTH_POOL_ID", "pool-xyz")
		t.Setenv("NUMBER", "42")
		t.Setenv("REQUIRED_KEY", "present")
		t.Setenv("SPEED", "42.5")
		t.Setenv("EXPIRE_TIME", "10")
		t.Setenv("COST", "100")
		t.Setenv("DEPLOY_ENV", "test")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Region != "us-east-1" {
			t.Errorf("Region = %q, want %q", cfg.Region, "us-east-1")
		}
		if cfg.CI != true {
			t.Errorf("CI = %v, want true", cfg.CI)
		}
		if cfg.Port != 27017 {
			t.Errorf("Port = %d, want 27017", cfg.Port)
		}
		if cfg.RequiredKey != "present" {
			t.Errorf("RequiredKey = %q, want %q", cfg.RequiredKey, "present")
		}
		if cfg.Pool.Value() != "pool-xyz" {
			t.Errorf("Pool.Value = %q, want %q", cfg.Pool.Value(), "pool-xyz")
		}
		if cfg.N.Value() != 42 {
			t.Errorf("N.Value = %d, want 42", cfg.N.Value())
		}
		if cfg.TTL != 30*time.Second {
			t.Errorf("TTL = %v, want 30s", cfg.TTL)
		}
		if cfg.Speed != 42.5 {
			t.Errorf("Speed = %f, want 42.5", cfg.Speed)
		}
		if cfg.ExpireTime != 10 {
			t.Errorf("ExpireTime = %v, want 10", cfg.ExpireTime)
		}
		if cfg.Cost != 100 {
			t.Errorf("Cost = %d, want 100", cfg.Cost)
		}
		if cfg.Ignored != "" {
			t.Errorf("Ignored should be empty, got %q", cfg.Ignored)
		}
		if cfg.private != "" {
			t.Errorf("private field must remain unset, got %q", cfg.private)
		}
	})

	t.Run("required missing errors", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")

		var cfg baseConfig
		err := Load(svc, &cfg)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !contains(err.Error(), "REQUIRED_KEY") {
			t.Errorf("error %q does not mention REQUIRED_KEY", err)
		}
	})

	t.Run("service precedence with default separator", func(t *testing.T) {
		svc := service.FromString("file")
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("file__AWS_REGION", "eu-west-1")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Region != "eu-west-1" {
			t.Errorf("Region precedence failed: got %q, want eu-west-1", cfg.Region)
		}
	})

	t.Run("custom separator", func(t *testing.T) {
		svc := service.FromString("app")
		t.Setenv("app::AWS_REGION", "ap-south-1")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg, Options{Separator: "::"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Region != "ap-south-1" {
			t.Errorf("Region = %q, want ap-south-1", cfg.Region)
		}
	})

	t.Run("bool parsing variants", func(t *testing.T) {
		t.Setenv("CI", "1")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.CI != true {
			t.Errorf("CI = %v, want true", cfg.CI)
		}
	})

	t.Run("Env wrapper captures global var name", func(t *testing.T) {
		t.Setenv("AUTH_POOL_ID", "global-pool")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Pool.VarName() != "AUTH_POOL_ID" {
			t.Errorf("VarName = %q, want AUTH_POOL_ID", cfg.Pool.VarName())
		}
	})

	t.Run("Env wrapper captures service-scoped var name", func(t *testing.T) {
		svc := service.FromString("svc")
		t.Setenv("svc__AUTH_POOL_ID", "svc-pool")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Pool.VarName() != "svc__AUTH_POOL_ID" {
			t.Errorf("VarName = %q, want svc__AUTH_POOL_ID", cfg.Pool.VarName())
		}
		if cfg.Pool.Value() != "svc-pool" {
			t.Errorf("Value = %q, want svc-pool", cfg.Pool.Value())
		}
	})

	t.Run("duration via TextUnmarshaler", func(t *testing.T) {
		t.Setenv("CACHE_TTL", "90s")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.TTL != 90*time.Second {
			t.Errorf("TTL = %v, want 90s", cfg.TTL)
		}
	})

	t.Run("skip tag ignores field even if set", func(t *testing.T) {
		t.Setenv("IGNORED", "should-not-be-set")
		t.Setenv("REQUIRED_KEY", "present")

		var cfg baseConfig
		if err := Load(svc, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Ignored != "" {
			t.Errorf("Ignored should remain empty, got %q", cfg.Ignored)
		}
	})

	t.Run("invalid tags cause errors", func(t *testing.T) {
		type bad struct {
			Bad1 string `env:""`
			Bad2 string `env:"X,unknown"`
		}

		var cfg bad
		err := Load(svc, &cfg)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !contains(err.Error(), "'env' tag cannot be empty") &&
			!contains(err.Error(), "unknown env tag attribute") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("target validation errors", func(t *testing.T) {
		var notPtr baseConfig
		if err := Load(svc, notPtr); err == nil {
			t.Errorf("expected error on non-pointer target")
		}

		var nilPtr *baseConfig
		if err := Load(svc, nilPtr); err == nil {
			t.Errorf("expected error on nil pointer")
		}

		type notStruct int
		var x notStruct
		if err := Load(svc, &x); err == nil {
			t.Errorf("expected error on non-struct target")
		}
	})
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}

	return indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}

	return -1
}
