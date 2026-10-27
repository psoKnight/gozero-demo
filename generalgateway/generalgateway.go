// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"

	"gozero-demo/common/healthing"
	"gozero-demo/common/resilience"
	"gozero-demo/generalgateway/internal/config"
	"gozero-demo/generalgateway/internal/handler"
	"gozero-demo/generalgateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/config_local.yaml", "the config file")

//go:embed docs/gateway.json
var openAPIDoc []byte

const swaggerPage = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>gozero-demo API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({url: "/api.json", dom_id: "#swagger-ui"})</script>
</body>
</html>`

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	limiter, err := resilience.NewRateLimiter(c.Resilience.HTTPRateLimit, "gateway-http")
	if err != nil {
		panic(fmt.Errorf("init HTTP rate limiter: %w", err))
	}
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(fmt.Errorf("init service context: %w", err))
	}
	handler.RegisterHandlers(server, ctx, limiter)
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api.json",
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write(openAPIDoc)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/swagger",
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(swaggerPage))
		},
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/status.html",
		Handler: healthing.Handler,
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
