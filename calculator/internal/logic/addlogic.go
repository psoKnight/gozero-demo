package logic

import (
	"context"
	"math"

	"github.com/zeromicro/go-zero/core/logx"
	"gozero-demo/calculator/internal/svc"
	"gozero-demo/pbdata/calculatorpb"
)

type AddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddLogic) Add(in *calculatorpb.BinaryOperationRequest) (*calculatorpb.OperationReply, error) {
	left, right := in.GetLeft(), in.GetRight()
	if (right > 0 && left > math.MaxInt64-right) ||
		(right < 0 && left < math.MinInt64-right) {
		return nil, overflowError()
	}

	return &calculatorpb.OperationReply{Result: left + right}, nil
}
