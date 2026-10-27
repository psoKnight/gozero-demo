// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"fmt"

	"gozero-demo/common/resilience"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Greeter    zrpc.RpcClientConf
	Calculator zrpc.RpcClientConf
	Resilience ResilienceConfig
}

type ResilienceConfig struct {
	HTTPRateLimit  resilience.RateLimitConfig `json:"httpRateLimit"`
	CircuitBreaker CircuitBreakerConfig       `json:"circuitBreaker"`
}

type CircuitBreakerConfig struct {
	Greeter    resilience.CircuitBreakerConfig `json:"greeter"`
	Calculator resilience.CircuitBreakerConfig `json:"calculator"`
}

func (c Config) Validate() error {
	if err := c.Resilience.HTTPRateLimit.Validate(); err != nil {
		return fmt.Errorf("resilience.httpRateLimit: %w", err)
	}
	if err := c.Resilience.CircuitBreaker.Greeter.Validate(); err != nil {
		return fmt.Errorf("resilience.circuitBreaker.greeter: %w", err)
	}
	if err := c.Resilience.CircuitBreaker.Calculator.Validate(); err != nil {
		return fmt.Errorf("resilience.circuitBreaker.calculator: %w", err)
	}
	return nil
}
