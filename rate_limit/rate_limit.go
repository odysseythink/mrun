package ratelimit

import "time"

// buildConfig combines defaults with options.
func buildConfig(opts []Option) config {
	c := config{
		slack: 10,
	}

	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

// config configures a limiter.
type config struct {
	slack int
}

type RateLimiter interface {
	// Take should block to make sure that the RPS is met.
	Take() time.Time
}
