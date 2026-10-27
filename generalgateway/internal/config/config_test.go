package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestConfigFiles(t *testing.T) {
	for _, file := range []string{"config_local.yaml", "config_test.yaml", "config.yaml"} {
		t.Run(file, func(t *testing.T) {
			var c Config
			if err := conf.Load("../../etc/"+file, &c); err != nil {
				t.Fatalf("conf.Load() error = %v", err)
			}
			if c.Port != 8080 {
				t.Fatalf("Port = %d, want 8080", c.Port)
			}
			if !c.Resilience.HTTPRateLimit.Enabled {
				t.Fatal("HTTP rate limit must be enabled")
			}
			if !c.Middlewares.Breaker || !c.Middlewares.Shedding {
				t.Fatal("REST breaker and shedding must be enabled")
			}
			if c.CpuThreshold <= 0 {
				t.Fatal("CPU threshold must be positive")
			}
			if !c.Greeter.Middlewares.Breaker || !c.Calculator.Middlewares.Breaker {
				t.Fatal("zrpc client breakers must be enabled")
			}
			if _, err := c.Greeter.BuildTarget(); err != nil {
				t.Fatalf("Greeter.BuildTarget() error = %v", err)
			}
			if _, err := c.Calculator.BuildTarget(); err != nil {
				t.Fatalf("Calculator.BuildTarget() error = %v", err)
			}
		})
	}
}
