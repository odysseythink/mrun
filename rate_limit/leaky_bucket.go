package ratelimit

import (
	"sync/atomic"
	"time"
)

type LeakyBucketRateLimit struct {
	state *atomic.Int64 // unix nanoseconds of the next permissions issue.

	perRequest time.Duration
	maxSlack   time.Duration
}

// NewLeakyBucketRateLimit returns a new atomic based limiter.
func NewLeakyBucketRateLimit(rate int, opts ...Option) *LeakyBucketRateLimit {
	// TODO consider moving config building to the implementation
	// independent code.
	config := buildConfig(opts)
	perRequest := time.Second / time.Duration(rate)
	l := &LeakyBucketRateLimit{
		perRequest: perRequest,
		maxSlack:   time.Duration(config.slack) * perRequest,
		state:      &atomic.Int64{},
	}
	l.state.Store(0)
	return l
}

// Take blocks to ensure that the time spent between multiple
// Take calls is on average time.Second/rate.
func (t *LeakyBucketRateLimit) Take() time.Time {
	var (
		newTimeOfNextPermissionIssue int64
		now                          int64
	)

	for {
		now = time.Now().UnixNano()
		timeOfNextPermissionIssue := t.state.Load()
		switch {
		case timeOfNextPermissionIssue == 0 || (t.maxSlack == 0 && now-timeOfNextPermissionIssue > int64(t.perRequest)):
			// if this is our first call or t.maxSlack == 0 we need to shrink issue time to now
			newTimeOfNextPermissionIssue = now
		case t.maxSlack > 0 && now-timeOfNextPermissionIssue > int64(t.maxSlack)+int64(t.perRequest):
			// a lot of nanoseconds passed since the last Take call
			// we will limit max accumulated time to maxSlack
			newTimeOfNextPermissionIssue = now - int64(t.maxSlack)
		default:
			// calculate the time at which our permission was issued
			newTimeOfNextPermissionIssue = timeOfNextPermissionIssue + int64(t.perRequest)
		}

		if t.state.CompareAndSwap(timeOfNextPermissionIssue, newTimeOfNextPermissionIssue) {
			break
		}
	}

	sleepDuration := time.Duration(newTimeOfNextPermissionIssue - now)
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
		return time.Unix(0, newTimeOfNextPermissionIssue)
	}
	// return now if we don't sleep as atomicLimiter does
	return time.Unix(0, now)
}
