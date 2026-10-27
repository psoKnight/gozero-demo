package main

import (
	"flag"
	"fmt"

	"gozero-demo/common/healthing"
	"gozero-demo/common/resilience"
	"gozero-demo/greeter/internal/config"
	"gozero-demo/greeter/internal/server"
	"gozero-demo/greeter/internal/svc"
	"gozero-demo/pbdata/greeterpb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/config_local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	limiter, err := resilience.NewRateLimiter(c.Resilience.GRPCRateLimit, "greeter-grpc")
	if err != nil {
		panic(fmt.Errorf("init gRPC rate limiter: %w", err))
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		greeterpb.RegisterGreeterServer(grpcServer, server.NewGreeterServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()
	if limiter != nil {
		s.AddUnaryInterceptors(limiter.UnaryServerInterceptor())
	}
	if err := healthing.Start(c.HealthServer.Address); err != nil {
		panic(fmt.Errorf("start health server: %w", err))
	}

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
