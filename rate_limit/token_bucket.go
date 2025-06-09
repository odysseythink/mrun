package ratelimit

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type TokenBucketRateLimit struct {
	// capacity holds the overall capacity of the bucket.
	capacity int64

	// quantum holds how many tokens are added on
	// each tick.
	quantum int64

	// fillInterval holds the interval between each tick.
	fillInterval time.Duration

	// availableTokens holds the number of available
	// tokens as of the associated latestTick.
	// It will be negative when there are consumers
	// waiting for tokens.
	availableTokens *atomic.Int64

	// latestTick holds the latest tick for which
	// we know the number of tokens in the bucket.
	latestTick *atomic.Int64

	// startTime holds the moment when the bucket was
	// first created and ticks began.
	startTime time.Time

	isStart sync.Once

	fillTimer *time.Timer

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
}

// NewBucket returns a new token bucket that fills at the
// rate of one token every fillInterval, up to the given
// maximum capacity. Both arguments must be
// positive. The bucket is initially full.
func NewTokenBucketRateLimit(fillInterval time.Duration, capacity int64) *TokenBucketRateLimit {
	rl := NewTokenBucketRateLimitWithQuantum(fillInterval, capacity, 1)
	rl.Take()
	return rl
}

// rateMargin specifes the allowed variance of actual
// rate from specified rate. 1% seems reasonable.
const rateMargin = 0.01

// NewBucketWithRate returns a token bucket that fills the bucket
// at the rate of rate tokens per second up to the given
// maximum capacity. Because of limited clock resolution,
// at high rates, the actual rate may be up to 1% different from the
// specified rate.
func NewTokenBucketRateLimitWithRate(rate float64, capacity int64) (*TokenBucketRateLimit, error) {
	// Use the same bucket each time through the loop
	// to save allocations.
	rl := &TokenBucketRateLimit{
		startTime:       time.Now(),
		latestTick:      &atomic.Int64{},
		fillInterval:    1,
		capacity:        capacity,
		quantum:         1,
		availableTokens: &atomic.Int64{},
	}
	rl.latestTick.Store(0)
	rl.availableTokens.Store(0)
	rl.ctx, rl.ctxCancelFunc = context.WithCancel(context.Background())
	for quantum := int64(1); quantum < 1<<50; quantum = nextQuantum(quantum) {
		fillInterval := time.Duration(float64(time.Second*time.Duration(quantum)) / rate)
		if fillInterval <= 0 {
			continue
		}
		rl.fillInterval = fillInterval
		rl.quantum = quantum
		if diff := math.Abs(rl.Rate() - rate); diff/rate <= rateMargin {
			rl.Take()
			return rl, nil
		}
	}
	return nil, fmt.Errorf("cannot find suitable quantum for %v", strconv.FormatFloat(rate, 'g', -1, 64))
}

// nextQuantum returns the next quantum to try after q.
// We grow the quantum exponentially, but slowly, so we
// get a good fit in the lower numbers.
func nextQuantum(q int64) int64 {
	q1 := q * 11 / 10
	if q1 == q {
		q1++
	}
	return q1
}

// NewBucketWithQuantum is similar to NewBucket, but allows
// the specification of the quantum size - quantum tokens
// are added every fillInterval.
func NewTokenBucketRateLimitWithQuantum(fillInterval time.Duration, capacity, quantum int64) *TokenBucketRateLimit {
	if fillInterval <= 0 {
		panic("token bucket fill interval is not > 0")
	}
	if capacity <= 0 {
		panic("token bucket capacity is not > 0")
	}
	if quantum <= 0 {
		panic("token bucket quantum is not > 0")
	}
	rl := &TokenBucketRateLimit{
		startTime:       time.Now(),
		latestTick:      &atomic.Int64{},
		fillInterval:    fillInterval,
		capacity:        capacity,
		quantum:         quantum,
		availableTokens: &atomic.Int64{},
	}
	rl.latestTick.Store(0)
	rl.availableTokens.Store(0)
	rl.ctx, rl.ctxCancelFunc = context.WithCancel(context.Background())
	rl.Take()
	return rl
}

func (tb *TokenBucketRateLimit) Take() time.Time {
	tb.isStart.Do(func() {
		go func() {
			tb.fillTimer = time.NewTimer(tb.fillInterval)
			for {
				select {
				case <-tb.fillTimer.C:
					available_tokens := tb.availableTokens.Load()
					if available_tokens+tb.quantum < tb.capacity {
						tb.availableTokens.Add(tb.quantum)
					}
					tb.fillTimer.Reset(tb.fillInterval)
				case <-tb.ctx.Done():
					return
				}
			}
		}()
	})
	available_tokens := tb.availableTokens.Add(-1)
	if available_tokens > 0 {
		now := time.Now().UnixNano()
		return time.Unix(0, now)
	} else {
		tb.availableTokens.Add(1)
		return time.Unix(0, 0)
	}
}

// Capacity returns the capacity that the bucket was created with.
func (tb *TokenBucketRateLimit) Capacity() int64 {
	return tb.capacity
}

// Rate returns the fill rate of the bucket, in tokens per second.
func (tb *TokenBucketRateLimit) Rate() float64 {
	return float64(time.Second*time.Duration(tb.quantum)) / float64(tb.fillInterval)
}
