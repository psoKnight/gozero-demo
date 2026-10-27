// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"fmt"

	"gozero-demo/generalgateway/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Greeter    zrpc.Client
	Calculator zrpc.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	greeter, err := newRPCClient(c.Greeter)
	if err != nil {
		return nil, fmt.Errorf("create greeter client: %w", err)
	}
	calculator, err := newRPCClient(c.Calculator)
	if err != nil {
		return nil, fmt.Errorf("create calculator client: %w", err)
	}
	return &ServiceContext{Config: c, Greeter: greeter, Calculator: calculator}, nil
}

func newRPCClient(conf zrpc.RpcClientConf) (zrpc.Client, error) {
	conf.Middlewares.Breaker = true
	return zrpc.NewClient(conf)
}
