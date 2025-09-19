package fleets

// Pool is a goroutine pool that limits and recycles a mass of goroutines.
// The pool capacity can be fixed or unlimited.
type Pool struct {
	*poolCommon
}

// Submit submits a task to the pool.
//
// Note that you are allowed to call Pool.Submit() from the current Pool.Submit(),
// but what calls for special attention is that you will get blocked with the last
// Pool.Submit() call once the current Pool runs out of its capacity, and to avoid this,
// you should instantiate a Pool with fleets.WithNonblocking(true).
func (p *Pool) Submit(task func()) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.inputFunc(task)
	}
	return err
}

// NewPool instantiates a Pool with customized options.
func NewPool(size int, options ...Option) (*Pool, error) {
	pc, err := newPool(size, options...)
	if err != nil {
		return nil, err
	}

	pool := &Pool{poolCommon: pc}
	pool.workerCache.New = func() any {
		return &goWorker{
			pool: pool,
			task: make(chan func(), workerChanCap),
		}
	}

	return pool, nil
}
