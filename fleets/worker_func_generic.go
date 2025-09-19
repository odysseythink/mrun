package fleets

import (
	"runtime/debug"
	"time"
)

// goWorkerWithFunc is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type goWorkerWithFuncGeneric[T any] struct {
	worker

	// pool who owns this worker.
	pool *PoolWithFuncGeneric[T]

	// arg is a job should be done.
	arg chan T

	// exit signals the goroutine to exit.
	exit chan struct{}

	// lastUsed will be updated when putting a worker back into queue.
	lastUsed time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *goWorkerWithFuncGeneric[T]) run() {
	w.pool.addRunning(1)
	go func() {
		defer func() {
			if w.pool.addRunning(-1) == 0 && w.pool.IsClosed() {
				w.pool.once.Do(func() {
					close(w.pool.allDone)
				})
			}
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.options.Logger.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
			w.pool.cond.Signal()
		}()

		for {
			select {
			case <-w.exit:
				return
			case arg := <-w.arg:
				w.pool.fn(arg)
				if ok := w.pool.revertWorker(w); !ok {
					return
				}
			}
		}
	}()
}

func (w *goWorkerWithFuncGeneric[T]) finish() {
	w.exit <- struct{}{}
}

func (w *goWorkerWithFuncGeneric[T]) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorkerWithFuncGeneric[T]) setLastUsedTime(t time.Time) {
	w.lastUsed = t
}
