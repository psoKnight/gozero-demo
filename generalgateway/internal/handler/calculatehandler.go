// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gozero-demo/generalgateway/internal/logic"
	"gozero-demo/generalgateway/internal/response"
	"gozero-demo/generalgateway/internal/svc"
	"gozero-demo/generalgateway/internal/types"
)

func CalculateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CalculateRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
			response.WriteError(w, r, response.ParamError, err.Error())
			return
		}

		l := logic.NewCalculateLogic(r.Context(), svcCtx)
		resp, err := l.Calculate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
