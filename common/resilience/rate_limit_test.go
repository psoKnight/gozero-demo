package resilience

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsBurstThenRejects(t *testing.T) {
	limiter, err := NewRateLimiter(RateLimitConfig{
		Enabled: true,
		Rate:    0.001,
		Burst:   1,
		KeyTTL:  time.Minute,
	}, "test")
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}
	if !limiter.allow("client") {
		t.Fatal("first request should be allowed")
	}
	if limiter.allow("client") {
		t.Fatal("second request should be rejected")
	}
	if !limiter.allow("another-client") {
		t.Fatal("different key should have an independent bucket")
	}
}
