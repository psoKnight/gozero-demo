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
	HTTPRateLimit resilience.RateLimitConfig `json:"httpRateLimit"`
}

func (c Config) Validate() error {
	if err := c.Resilience.HTTPRateLimit.Validate(); err != nil {
		return fmt.Errorf("resilience.httpRateLimit: %w", err)
	}
	return nil
}
