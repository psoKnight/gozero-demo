package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	Success       = 0
	ParamError    = 1300
	InternalError = 1301
	RateLimited   = 1429
	Unavailable   = 1503
)

var messages = map[int]string{
	Success:       "成功",
	ParamError:    "参数错误",
	InternalError: "内部错误",
	RateLimited:   "请求过于频繁，请稍后重试",
	Unavailable:   "下游服务暂时不可用，请稍后重试",
}

type errorResponse struct {
	ErrorNo  int    `json:"error_no"`
	ErrorMsg string `json:"error_msg"`
}

// Message returns the standard message for a business code.
func Message(code int) string {
	return messages[code]
}

// WriteError writes a unified business error while keeping HTTP 200 semantics.
func WriteError(w http.ResponseWriter, r *http.Request, code int, message string) {
	if message == "" {
		message = Message(code)
	}
	httpx.OkJsonCtx(r.Context(), w, errorResponse{
		ErrorNo:  code,
		ErrorMsg: message,
	})
}

// UpstreamError maps a downstream gRPC status to the HTTP business error contract.
func UpstreamError(err error) (int, string) {
	grpcStatus := status.Convert(err)
	switch grpcStatus.Code() {
	case codes.InvalidArgument, codes.OutOfRange:
		return ParamError, grpcStatus.Message()
	case codes.ResourceExhausted:
		return RateLimited, Message(RateLimited)
	case codes.Unavailable:
		return Unavailable, Message(Unavailable)
	default:
		return InternalError, err.Error()
	}
}
