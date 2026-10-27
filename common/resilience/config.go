package resilience

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// RateLimitConfig configures go-zero Redis token buckets.
type RateLimitConfig struct {
	Enabled bool            `json:"enabled,default=false"`
	Rate    int             `json:"rate,optional"`
	Burst   int             `json:"burst,optional"`
	KeyTTL  time.Duration   `json:"keyTTL,default=10m"`
	Redis   redis.RedisConf `json:"redis,optional"`
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
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	return nil
}
