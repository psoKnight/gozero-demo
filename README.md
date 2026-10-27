# gozero-demo

使用 Go-Zero 完整实现 `goframe-demo` 的 HTTP + gRPC 微服务示例。

## 架构

```text
HTTP request
  -> generalgateway (go-zero rest)
  -> greeter / calculator (go-zero zrpc)
```

- `generalgateway`：HTTP JSON 网关，提供 `POST /v1/hello` 和 `POST /v1/calculate`。
- `greeter`：gRPC 问候服务，监听 `50051`，健康检查监听 `50052`。
- `calculator`：gRPC 整数计算服务，监听 `50053`，健康检查监听 `50054`。
- `common/resilience`：进程内 Key 级令牌桶、gRPC 服务端限流和分下游熔断器。
- `pbdata`：protobuf 协议及生成代码。
- `generalgateway/gateway.api`：go-zero API 路由定义。
- `generalgateway/docs/gateway.json`：OpenAPI 2.0 文档。

## 本地运行

要求 Go 1.26+、goctl 1.10+、Protocol Buffers、`protoc-gen-go` 和
`protoc-gen-go-grpc`。

分别在三个终端启动：

```bash
go run ./greeter -f=greeter/etc/config_local.yaml
go run ./calculator -f=calculator/etc/config_local.yaml
go run ./generalgateway -f=generalgateway/etc/config_local.yaml
```

网关入口为 `http://127.0.0.1:8080`。本地配置使用 gRPC 直连，网关指向
`127.0.0.1:50051` 和 `127.0.0.1:50053`，不依赖 Etcd。

## API

问候：

```bash
curl -X POST http://localhost:8080/v1/hello \
  -H "Content-Type: application/json" \
  -d '{"name":"Codex"}'
```

```json
{"error_no":0,"error_msg":"成功","data":{"message":"Hello, Codex! (Greeter)"}}
```

加法：

```bash
curl -X POST http://localhost:8080/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"add","left":21,"right":21}'
```

```json
{"error_no":0,"error_msg":"成功","data":{"result":42}}
```

乘法：

```bash
curl -X POST http://localhost:8080/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"multiply","left":"6","right":"7"}'
```

`left` 和 `right` 接受 JSON 整数或十进制字符串。字符串形式可避免客户端在解析
大整数时丢失精度。超出 `int64` 范围时返回：

```json
{"error_no":1300,"error_msg":"integer overflow"}
```

业务错误码：

| error_no | 含义 |
| --- | --- |
| 0 | 成功 |
| 1300 | 参数错误、下游参数错误或整数溢出 |
| 1301 | 内部错误 |
| 1429 | 请求过于频繁 |
| 1503 | 下游服务暂时不可用或熔断已打开 |

健康检查与文档：

```text
GET http://localhost:50052/status.html
GET http://localhost:50054/status.html
GET http://localhost:8080/status.html
GET http://localhost:8080/api.json
GET http://localhost:8080/swagger
```

## 服务发现

三套配置：

| 文件 | 用途 | 发现方式 |
| --- | --- | --- |
| `config_local.yaml` | 本机开发 | Gateway 直连两个 gRPC 服务 |
| `config_test.yaml` | Compose / 测试环境 | Etcd `etcd:2379` |
| `config.yaml` | 正式环境 | Etcd 占位地址，发布前必须替换 |

测试和正式配置中的服务注册键为：

```text
greeter.rpc
calculator.rpc
```

Go-Zero zrpc 内置 gRPC Health 服务并注册到 Etcd。进程异常退出后，Etcd 租约自动
过期，客户端通过 Watch 更新端点。

## 限流与熔断

HTTP 限流维度为客户端 IP + HTTP Method + Path，超过限制时返回 HTTP 429：

```json
{"error_no":1429,"error_msg":"请求过于频繁，请稍后重试"}
```

Greeter 和 Calculator 在 gRPC 服务端按方法分别限流，超限使用标准 gRPC
`ResourceExhausted`。网关会把它转换成业务码 1429。

Gateway 为 Greeter 和 Calculator 分别维护熔断器。`InvalidArgument`、
`OutOfRange`、`ResourceExhausted` 等业务状态不会计入系统失败；超时、连接失败、
`Unavailable` 等错误会累计失败次数。熔断打开后，网关直接返回业务码 1503。

所有下游调用同时有 3 秒上下文超时和 zrpc 客户端超时。限流器与熔断器均为进程级
状态；如果要求集群级统一配额，应在 API Gateway、Service Mesh 或 Redis 中实现。

## 新增 Calculator 运算

以新增 `Subtract`（减法）为例，一个可调用的 Calculator 运算需要同步修改
protobuf、服务端 Logic、gRPC Server、客户端封装和 Gateway 分发。

### 1. 修改 protobuf

在 `pbdata/calculatorpb/calculator.proto` 的 `Calculator` service 中增加 RPC：

```proto
service Calculator {
  rpc Add(BinaryOperationRequest) returns (OperationReply);
  rpc Multiply(BinaryOperationRequest) returns (OperationReply);
  rpc Subtract(BinaryOperationRequest) returns (OperationReply);
}
```

重新生成 protobuf：

```bash
bash pbdata/calculatorpb/pb.sh
```

### 2. 新增业务 Logic

新增 `calculator/internal/logic/subtractlogic.go`：

```go
package logic

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	"gozero-demo/calculator/internal/svc"
	"gozero-demo/pbdata/calculatorpb"
)

type SubtractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubtractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubtractLogic {
	return &SubtractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubtractLogic) Subtract(in *calculatorpb.BinaryOperationRequest) (*calculatorpb.OperationReply, error) {
	result := new(big.Int).Sub(
		big.NewInt(in.GetLeft()),
		big.NewInt(in.GetRight()),
	)
	if !result.IsInt64() {
		return nil, overflowError()
	}
	return &calculatorpb.OperationReply{Result: result.Int64()}, nil
}
```

### 3. 挂载 gRPC Server

`calculator/internal/server/calculatorserver.go` 是 goctl 生成文件，不建议直接
修改。新增 `calculator/internal/server/subtractserver.go`：

```go
package server

import (
	"context"

	"gozero-demo/calculator/internal/logic"
	"gozero-demo/pbdata/calculatorpb"
)

func (s *CalculatorServer) Subtract(
	ctx context.Context,
	in *calculatorpb.BinaryOperationRequest,
) (*calculatorpb.OperationReply, error) {
	return logic.NewSubtractLogic(ctx, s.svcCtx).Subtract(in)
}
```

### 4. 更新 gRPC 客户端封装

在 `calculator/calculator/calculator.go` 的 `Calculator` interface 中加入：

```go
Subtract(ctx context.Context, in *BinaryOperationRequest, opts ...grpc.CallOption) (*OperationReply, error)
```

并增加实现：

```go
func (m *defaultCalculator) Subtract(
	ctx context.Context,
	in *BinaryOperationRequest,
	opts ...grpc.CallOption,
) (*OperationReply, error) {
	client := calculatorpb.NewCalculatorClient(m.cli.Conn())
	return client.Subtract(ctx, in, opts...)
}
```

### 5. 接入 Gateway

如果该运算继续复用 `POST /v1/calculate`，修改
`generalgateway/internal/logic/calculatelogic.go`：

```go
if operation != "add" && operation != "multiply" && operation != "subtract" {
	return calculateError(response.ParamError, "operation 仅支持 add、multiply 或 subtract"), nil
}

switch operation {
case "add":
	reply, err = client.Add(ctx, callReq)
case "multiply":
	reply, err = client.Multiply(ctx, callReq)
case "subtract":
	reply, err = client.Subtract(ctx, callReq)
}
```

调用示例：

```bash
curl -X POST http://localhost:8080/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"subtract","left":50,"right":8}'
```

```json
{"error_no":0,"error_msg":"成功","data":{"result":42}}
```

### 6. 补充测试和文档

- 在 `calculator/internal/logic` 增加正常值、边界值和溢出测试。
- 在 `generalgateway/internal/logic` 增加新 `operation` 的分发测试。
- 同步更新 `generalgateway/docs/gateway.json` 中 `operation` 的说明。
- 执行 `gofmt`、`go test ./...`、`go vet ./...`。

如果新增的是完全独立的新 HTTP 接口，还需要同步修改
`generalgateway/gateway.api`，再生成或补充对应 handler、logic 和 API types。

## 协议与代码生成

重新生成 protobuf：

```bash
bash pbdata/greeterpb/pb.sh
bash pbdata/calculatorpb/pb.sh
```

重新生成 API 骨架、校验 API 定义和 OpenAPI：

```bash
goctl api go --api generalgateway/gateway.api --dir generalgateway --style gozero
goctl api validate --api generalgateway/gateway.api
goctl api swagger --api generalgateway/gateway.api --dir generalgateway/docs --filename gateway
```

`goctl api go` 只应生成缺失文件；再次生成前请检查它不会覆盖已手写逻辑。生成
OpenAPI 后，需要把 `left/right` 的 schema 保持为整数或字符串二选一。

## 验证

```bash
go test ./...
go vet ./...
goctl api validate --api generalgateway/gateway.api
```

## Docker Compose

先为各服务生成 Linux 发布目录：

```bash
(cd greeter && bash deploy.sh -b test)
(cd calculator && bash deploy.sh -b test)
(cd generalgateway && bash deploy.sh -b test)
docker compose up --build -d
```

Compose 内置单节点 Etcd，Greeter 和 Calculator 不占用宿主固定端口。可动态扩缩容：

```bash
docker compose up -d --scale greeter=3 --scale calculator=2
```

发布镜像：

```bash
cd greeter
IMAGE_REPO=harbor.example.com/gozero-demo bash deploy.sh -d 1.0.0 test
```

Calculator 和 Gateway 使用同样方式发布。正式环境必须替换三套
`config.yaml` 中的 Etcd 占位地址，并将明文 gRPC 替换为安全传输方案。
