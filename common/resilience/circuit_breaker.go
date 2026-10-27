package resilience

import (
	"context"
	"errors"

	"github.com/sony/gobreaker/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientBreaker wraps one downstream gRPC service with a circuit breaker.
type ClientBreaker struct {
	name    string
	breaker *gobreaker.CircuitBreaker[any]
}

// NewClientBreaker creates a breaker when the supplied config is enabled.
func NewClientBreaker(conf CircuitBreakerConfig, name string) (*ClientBreaker, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if !conf.Enabled {
		return nil, nil
	}

	return &ClientBreaker{
		name: name,
		breaker: gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
			Name:        name,
			MaxRequests: conf.HalfOpenMaxRequests,
			Interval:    conf.Interval,
			Timeout:     conf.Timeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= conf.FailureThreshold
			},
			IsExcluded: isBusinessError,
			OnStateChange: func(name string, from, to gobreaker.State) {
				logx.Infof("circuit breaker state changed: %s %s -> %s", name, from, to)
			},
		}),
	}, nil
}

// UnaryClientInterceptor executes the downstream RPC through the breaker.
func (b *ClientBreaker) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := b.breaker.Execute(func() (any, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return status.Errorf(codes.Unavailable, "%s service is temporarily unavailable", b.name)
		}
		return err
	}
}

func isBusinessError(err error) bool {
	if err == nil {
		return false
	}
	switch status.Code(err) {
	case codes.Canceled,
		codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.OutOfRange,
		codes.FailedPrecondition,
		codes.ResourceExhausted:
		return true
	default:
		return false
	}
}
