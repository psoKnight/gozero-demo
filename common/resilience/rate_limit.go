package resilience

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const rateLimitedMessage = "请求过于频繁，请稍后重试"

type limiterEntry struct {
	limiter  *limit.TokenLimiter
	lastSeen time.Time
}

type rateLimitResponse struct {
	ErrorNo  int    `json:"error_no"`
	ErrorMsg string `json:"error_msg"`
}

// RateLimiter provides independent in-process token buckets per key.
type RateLimiter struct {
	name        string
	rate        int
	burst       int
	keyTTL      time.Duration
	mu          sync.Mutex
	entries     map[string]*limiterEntry
	lastCleanup time.Time
	store       *redis.Redis
}

// NewRateLimiter creates a limiter when the supplied config is enabled.
func NewRateLimiter(conf RateLimitConfig, name string) (*RateLimiter, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if !conf.Enabled {
		return nil, nil
	}
	store, err := redis.NewRedis(conf.Redis)
	if err != nil {
		return nil, fmt.Errorf("create rate limit redis: %w", err)
	}
	return &RateLimiter{
		name:    name,
		rate:    conf.Rate,
		burst:   conf.Burst,
		keyTTL:  conf.KeyTTL,
		entries: make(map[string]*limiterEntry),
		store:   store,
	}, nil
}

// HTTPMiddleware limits requests by client IP, HTTP method, and path.
func (l *RateLimiter) HTTPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprintf("http|%s|%s|%s", clientIP(r), r.Method, r.URL.Path)
		if l.allow(r.Context(), key) {
			next(w, r)
			return
		}

		logx.WithContext(r.Context()).Errorf("rate limit rejected: limiter=%s key=%s", l.name, key)
		body, _ := json.Marshal(rateLimitResponse{
			ErrorNo:  1429,
			ErrorMsg: rateLimitedMessage,
		})
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write(body)
	}
}

// UnaryServerInterceptor applies a separate limit to each RPC method.
func (l *RateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		key := "grpc|" + info.FullMethod
		if !l.allow(ctx, key) {
			logx.WithContext(ctx).Errorf("rate limit rejected: limiter=%s key=%s", l.name, key)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func (l *RateLimiter) allow(ctx context.Context, key string) bool {
	now := time.Now()
	l.mu.Lock()
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= l.keyTTL {
		for entryKey, entry := range l.entries {
			if now.Sub(entry.lastSeen) >= l.keyTTL {
				delete(l.entries, entryKey)
			}
		}
		l.lastCleanup = now
	}

	entry, ok := l.entries[key]
	if !ok {
		entry = &limiterEntry{limiter: limit.NewTokenLimiter(l.rate, l.burst, l.store, l.name+"|"+key)}
		l.entries[key] = entry
	}
	entry.lastSeen = now
	limiter := entry.limiter
	l.mu.Unlock()

	return limiter.AllowCtx(ctx)
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
			return first
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
