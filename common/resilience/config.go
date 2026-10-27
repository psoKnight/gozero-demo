package resilience

import (
	"fmt"
	"time"
)

// RateLimitConfig configures a local keyed token bucket.
type RateLimitConfig struct {
	Enabled bool          `json:"enabled,default=false"`
	Rate    float64       `json:"rate,optional"`
	Burst   int           `json:"burst,optional"`
	KeyTTL  time.Duration `json:"keyTTL,default=10m"`
}

// Validate checks that an enabled limiter has usable parameters.
func (c RateLimitConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Rate <= 0 {
		return fmt.Errorf("rate must be greater than 0")
	}
	if c.Burst <= 0 {
		return fmt.Errorf("burst must be greater than 0")
	}
	if c.KeyTTL <= 0 {
		return fmt.Errorf("keyTTL must be greater than 0")
	}
	return nil
}

// CircuitBreakerConfig configures an outbound gRPC circuit breaker.
type CircuitBreakerConfig struct {
	Enabled             bool          `json:"enabled,default=false"`
	FailureThreshold    uint32        `json:"failureThreshold,optional"`
	Interval            time.Duration `json:"interval,optional"`
	Timeout             time.Duration `json:"timeout,optional"`
	HalfOpenMaxRequests uint32        `json:"halfOpenMaxRequests,default=1"`
}

// Validate checks that an enabled breaker has usable parameters.
func (c CircuitBreakerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.FailureThreshold == 0 {
		return fmt.Errorf("failureThreshold must be greater than 0")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	if c.HalfOpenMaxRequests == 0 {
		return fmt.Errorf("halfOpenMaxRequests must be greater than 0")
	}
	return nil
}
