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

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const rateLimitedMessage = "请求过于频繁，请稍后重试"

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimitResponse struct {
	ErrorNo  int    `json:"error_no"`
	ErrorMsg string `json:"error_msg"`
}

// RateLimiter provides independent in-process token buckets per key.
type RateLimiter struct {
	name        string
	limit       rate.Limit
	burst       int
	keyTTL      time.Duration
	mu          sync.Mutex
	entries     map[string]*limiterEntry
	lastCleanup time.Time
}

// NewRateLimiter creates a limiter when the supplied config is enabled.
func NewRateLimiter(conf RateLimitConfig, name string) (*RateLimiter, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if !conf.Enabled {
		return nil, nil
	}
	return &RateLimiter{
		name:    name,
		limit:   rate.Limit(conf.Rate),
		burst:   conf.Burst,
		keyTTL:  conf.KeyTTL,
		entries: make(map[string]*limiterEntry),
	}, nil
}

// HTTPMiddleware limits requests by client IP, HTTP method, and path.
func (l *RateLimiter) HTTPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprintf("http|%s|%s|%s", clientIP(r), r.Method, r.URL.Path)
		if l.allow(key) {
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
		if !l.allow(key) {
			logx.WithContext(ctx).Errorf("rate limit rejected: limiter=%s key=%s", l.name, key)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func (l *RateLimiter) allow(key string) bool {
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
		entry = &limiterEntry{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.entries[key] = entry
	}
	entry.lastSeen = now
	limiter := entry.limiter
	l.mu.Unlock()

	return limiter.AllowN(now, 1)
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
