package fleets

import (
	"runtime/debug"
	"time"
)

// goWorkerWithFunc is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type goWorkerWithFunc struct {
	worker

	// pool who owns this worker.
	pool *PoolWithFunc

	// arg is the argument for the function.
	arg chan any

	// lastUsed will be updated when putting a worker back into queue.
	lastUsed time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *goWorkerWithFunc) run() {
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

		for arg := range w.arg {
			if arg == nil {
				return
			}
			w.pool.fn(arg)
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}

func (w *goWorkerWithFunc) finish() {
	w.arg <- nil
}

func (w *goWorkerWithFunc) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorkerWithFunc) setLastUsedTime(t time.Time) {
	w.lastUsed = t
}

func (w *goWorkerWithFunc) inputArg(arg any) {
	w.arg <- arg
}
