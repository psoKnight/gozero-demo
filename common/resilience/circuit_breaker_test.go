package resilience

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCircuitBreakerOpensAndRejects(t *testing.T) {
	breaker, err := NewClientBreaker(CircuitBreakerConfig{
		Enabled:             true,
		FailureThreshold:    1,
		Interval:            time.Minute,
		Timeout:             time.Minute,
		HalfOpenMaxRequests: 1,
	}, "test")
	if err != nil {
		t.Fatalf("NewClientBreaker() error = %v", err)
	}
	calls := 0
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls++
		return status.Error(codes.Unavailable, "downstream unavailable")
	}
	interceptor := breaker.UnaryClientInterceptor()

	_ = interceptor(context.Background(), "/test", nil, nil, nil, invoker)
	err = interceptor(context.Background(), "/test", nil, nil, nil, invoker)
	if got := status.Code(err); got != codes.Unavailable {
		t.Fatalf("second call code = %v, want %v", got, codes.Unavailable)
	}
	if calls != 1 {
		t.Fatalf("downstream calls = %d, want 1", calls)
	}
}

func TestCircuitBreakerExcludesBusinessErrors(t *testing.T) {
	breaker, err := NewClientBreaker(CircuitBreakerConfig{
		Enabled:             true,
		FailureThreshold:    1,
		Interval:            time.Minute,
		Timeout:             time.Minute,
		HalfOpenMaxRequests: 1,
	}, "test")
	if err != nil {
		t.Fatalf("NewClientBreaker() error = %v", err)
	}
	calls := 0
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls++
		return status.Error(codes.InvalidArgument, "invalid request")
	}
	interceptor := breaker.UnaryClientInterceptor()

	for range 3 {
		_ = interceptor(context.Background(), "/test", nil, nil, nil, invoker)
	}
	if calls != 3 {
		t.Fatalf("downstream calls = %d, want 3", calls)
	}
}
