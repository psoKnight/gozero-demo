package resilience

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRateLimiterAllowsBurstThenRejects(t *testing.T) {
	mr := miniredis.RunT(t)
	limiter, err := NewRateLimiter(RateLimitConfig{
		Enabled: true,
		Rate:    1,
		Burst:   1,
		KeyTTL:  time.Minute,
		Redis: redis.RedisConf{
			Host: mr.Addr(),
			Type: redis.NodeType,
		},
	}, "test")
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}
	if !limiter.allow(context.Background(), "client") {
		t.Fatal("first request should be allowed")
	}
	if limiter.allow(context.Background(), "client") {
		t.Fatal("second request should be rejected")
	}
	if !limiter.allow(context.Background(), "another-client") {
		t.Fatal("different key should have an independent bucket")
	}
}

func TestHTTPRateLimitMiddleware(t *testing.T) {
	mr := miniredis.RunT(t)
	limiter, err := NewRateLimiter(RateLimitConfig{
		Enabled: true,
		Rate:    1,
		Burst:   1,
		KeyTTL:  time.Minute,
		Redis: redis.RedisConf{
			Host: mr.Addr(),
			Type: redis.NodeType,
		},
	}, "test")
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}
	handler := limiter.HTTPMiddleware(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	first := httptest.NewRecorder()
	handler(first, httptest.NewRequest(http.MethodPost, "/v1/test", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", first.Code, http.StatusNoContent)
	}

	second := httptest.NewRecorder()
	handler(second, httptest.NewRequest(http.MethodPost, "/v1/test", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
}

func TestGRPCRateLimitInterceptor(t *testing.T) {
	mr := miniredis.RunT(t)
	limiter, err := NewRateLimiter(RateLimitConfig{
		Enabled: true,
		Rate:    1,
		Burst:   1,
		KeyTTL:  time.Minute,
		Redis: redis.RedisConf{
			Host: mr.Addr(),
			Type: redis.NodeType,
		},
	}, "test")
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}

	interceptor := limiter.UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Call"}
	handler := func(context.Context, any) (any, error) {
		return "ok", nil
	}

	if _, err := interceptor(context.Background(), nil, info, handler); err != nil {
		t.Fatalf("first call error = %v", err)
	}
	if _, err := interceptor(context.Background(), nil, info, handler); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("second call code = %v, want %v", status.Code(err), codes.ResourceExhausted)
	}
}
