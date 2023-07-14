package mrun

import "time"

// WorkerOption represents the optional function.
type WorkerOption func(opts *WorkerOptions)

func loadWorkerOptions(options ...WorkerOption) *WorkerOptions {
	opts := new(WorkerOptions)
	for _, option := range options {
		option(opts)
	}
	return opts
}

// WorkerOptions contains all options which will be applied when instantiating an ants pool.
type WorkerOptions struct {
	// ExpiryDuration is a period for the scavenger goroutine to clean up those expired workers,
	// the scavenger scans all workers every `ExpiryDuration` and clean up those workers that haven't been
	// used for more than `ExpiryDuration`.
	ExpiryDuration time.Duration

	// PreAlloc indicates whether to make memory pre-allocation when initializing Pool.
	PreAlloc bool

	// Max number of goroutine blocking on pool.Submit.
	// 0 (default value) means no such limit.
	MaxBlockingTasks int

	// When Nonblocking is true, Pool.Submit will never be blocked.
	// ErrPoolOverload will be returned when Pool.Submit cannot be done at once.
	// When Nonblocking is true, MaxBlockingTasks is inoperative.
	Nonblocking bool

	// PanicHandler is used to handle panics from each worker goroutine.
	// if nil, panics will be thrown out again from worker goroutines.
	PanicHandler func(interface{})

	// When DisablePurge is true, workers are not purged and are resident.
	DisablePurge bool
}

// WithOptions accepts the whole options config.
func WithWorkerOptions(options WorkerOptions) WorkerOption {
	return func(opts *WorkerOptions) {
		*opts = options
	}
}

// WithWorkerExpiryDuration sets up the interval time of cleaning up goroutines.
func WithWorkerExpiryDuration(expiryDuration time.Duration) WorkerOption {
	return func(opts *WorkerOptions) {
		opts.ExpiryDuration = expiryDuration
	}
}

// WithWorkerPreAlloc indicates whether it should malloc for workers.
func WithWorkerPreAlloc(preAlloc bool) WorkerOption {
	return func(opts *WorkerOptions) {
		opts.PreAlloc = preAlloc
	}
}

// WithWorkerMaxBlockingTasks sets up the maximum number of goroutines that are blocked when it reaches the capacity of pool.
func WithWorkerMaxBlockingTasks(maxBlockingTasks int) WorkerOption {
	return func(opts *WorkerOptions) {
		opts.MaxBlockingTasks = maxBlockingTasks
	}
}

// WithWorkerNonblocking indicates that pool will return nil when there is no available workers.
func WithWorkerNonblocking(nonblocking bool) WorkerOption {
	return func(opts *WorkerOptions) {
		opts.Nonblocking = nonblocking
	}
}

// WithWorkerPanicHandler sets up panic handler.
func WithWorkerPanicHandler(panicHandler func(interface{})) WorkerOption {
	return func(opts *WorkerOptions) {
		opts.PanicHandler = panicHandler
	}
}

// WithWorkerDisablePurge indicates whether we turn off automatically purge.
func WithWorkerDisablePurge(disable bool) WorkerOption {
	return func(opts *WorkerOptions) {
		opts.DisablePurge = disable
	}
}
