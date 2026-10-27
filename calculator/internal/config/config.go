package config

import (
	"fmt"

	"gozero-demo/common/resilience"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	HealthServer HealthServerConfig
	Resilience   ResilienceConfig
}

type HealthServerConfig struct {
	Address string
}

type ResilienceConfig struct {
	GRPCRateLimit resilience.RateLimitConfig `json:"grpcRateLimit"`
}

func (c Config) Validate() error {
	if c.HealthServer.Address == "" {
		return fmt.Errorf("healthServer.address is required")
	}
	if err := c.Resilience.GRPCRateLimit.Validate(); err != nil {
		return fmt.Errorf("resilience.grpcRateLimit: %w", err)
	}
	return nil
}
