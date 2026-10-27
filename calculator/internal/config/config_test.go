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
			if c.Name != "calculator.rpc" {
				t.Fatalf("Name = %q, want calculator.rpc", c.Name)
			}
			if c.HealthServer.Address == "" {
				t.Fatal("HealthServer.Address is empty")
			}
			if !c.Resilience.GRPCRateLimit.Enabled {
				t.Fatal("gRPC rate limit must be enabled")
			}
		})
	}
}
