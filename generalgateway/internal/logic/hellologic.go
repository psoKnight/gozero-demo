// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"strings"
	"time"

	"gozero-demo/generalgateway/internal/response"
	"gozero-demo/generalgateway/internal/svc"
	"gozero-demo/generalgateway/internal/types"
	greeterclient "gozero-demo/greeter/greeter"
	"gozero-demo/pbdata/greeterpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HelloLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHelloLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HelloLogic {
	return &HelloLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HelloLogic) Hello(req *types.HelloRequest) (resp *types.HelloResponse, err error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return &types.HelloResponse{
			ErrorNo:  response.ParamError,
			ErrorMsg: "name 不能为空",
		}, nil
	}

	ctx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
	defer cancel()
	reply, err := greeterclient.NewGreeter(l.svcCtx.Greeter).SayHello(ctx, &greeterpb.HelloRequest{
		Name: name,
	})
	if err != nil {
		code, message := response.UpstreamError(err)
		return &types.HelloResponse{ErrorNo: code, ErrorMsg: message}, nil
	}

	return &types.HelloResponse{
		ErrorNo:  response.Success,
		ErrorMsg: response.Message(response.Success),
		Data:     &types.HelloData{Message: reply.GetMessage()},
	}, nil
}
