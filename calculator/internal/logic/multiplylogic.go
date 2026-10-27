package logic

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	"gozero-demo/calculator/internal/svc"
	"gozero-demo/pbdata/calculatorpb"
)

type MultiplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMultiplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MultiplyLogic {
	return &MultiplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MultiplyLogic) Multiply(in *calculatorpb.BinaryOperationRequest) (*calculatorpb.OperationReply, error) {
	result := new(big.Int).Mul(big.NewInt(in.GetLeft()), big.NewInt(in.GetRight()))
	if !result.IsInt64() {
		return nil, overflowError()
	}

	return &calculatorpb.OperationReply{Result: result.Int64()}, nil
}
