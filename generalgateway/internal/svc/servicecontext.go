// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"fmt"

	"gozero-demo/common/resilience"
	"gozero-demo/generalgateway/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Greeter    zrpc.Client
	Calculator zrpc.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	greeter, err := newRPCClient(c.Greeter, c.Resilience.CircuitBreaker.Greeter, "greeter")
	if err != nil {
		return nil, fmt.Errorf("create greeter client: %w", err)
	}
	calculator, err := newRPCClient(c.Calculator, c.Resilience.CircuitBreaker.Calculator, "calculator")
	if err != nil {
		return nil, fmt.Errorf("create calculator client: %w", err)
	}
	return &ServiceContext{Config: c, Greeter: greeter, Calculator: calculator}, nil
}

func newRPCClient(conf zrpc.RpcClientConf, breakerConf resilience.CircuitBreakerConfig, name string) (zrpc.Client, error) {
	conf.Middlewares.Breaker = false
	breaker, err := resilience.NewClientBreaker(breakerConf, name)
	if err != nil {
		return nil, err
	}
	var options []zrpc.ClientOption
	if breaker != nil {
		options = append(options, zrpc.WithUnaryClientInterceptor(breaker.UnaryClientInterceptor()))
	}
	return zrpc.NewClient(conf, options...)
}
