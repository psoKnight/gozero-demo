// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	calculatorclient "gozero-demo/calculator/calculator"
	"gozero-demo/generalgateway/internal/response"
	"gozero-demo/generalgateway/internal/svc"
	"gozero-demo/generalgateway/internal/types"
	"gozero-demo/pbdata/calculatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CalculateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCalculateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalculateLogic {
	return &CalculateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CalculateLogic) Calculate(req *types.CalculateRequest) (resp *types.CalculateResponse, err error) {
	operation := strings.ToLower(strings.TrimSpace(req.Operation))
	if operation == "" {
		return calculateError(response.ParamError, "operation 不能为空"), nil
	}
	if operation != "add" && operation != "multiply" {
		return calculateError(response.ParamError, "operation 仅支持 add 或 multiply"), nil
	}
	left, err := parseInt64Field(req.Left)
	if err != nil {
		return calculateError(response.ParamError, "left 必须是 int64 整数"), nil
	}
	right, err := parseInt64Field(req.Right)
	if err != nil {
		return calculateError(response.ParamError, "right 必须是 int64 整数"), nil
	}

	ctx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
	defer cancel()
	client := calculatorclient.NewCalculator(l.svcCtx.Calculator)
	callReq := &calculatorpb.BinaryOperationRequest{Left: left, Right: right}
	var reply *calculatorpb.OperationReply
	if operation == "add" {
		reply, err = client.Add(ctx, callReq)
	} else {
		reply, err = client.Multiply(ctx, callReq)
	}
	if err != nil {
		code, message := response.UpstreamError(err)
		return calculateError(code, message), nil
	}

	return &types.CalculateResponse{
		ErrorNo:  response.Success,
		ErrorMsg: response.Message(response.Success),
		Data:     &types.CalculateData{Result: reply.GetResult()},
	}, nil
}

func parseInt64Field(raw json.RawMessage) (int64, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" {
		return 0, strconv.ErrSyntax
	}
	if value[0] == '"' {
		if err := json.Unmarshal(raw, &value); err != nil {
			return 0, err
		}
	}
	return strconv.ParseInt(strings.TrimSpace(value), 10, 64)
}

func calculateError(code int, message string) *types.CalculateResponse {
	return &types.CalculateResponse{ErrorNo: code, ErrorMsg: message}
}
