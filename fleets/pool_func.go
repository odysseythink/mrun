package fleets

// PoolWithFunc is like Pool but accepts a unified function for all goroutines to execute.
type PoolWithFunc struct {
	*poolCommon

	// fn is the unified function for processing tasks.
	fn func(any)
}

// Invoke passes arguments to the pool.
//
// Note that you are allowed to call Pool.Invoke() from the current Pool.Invoke(),
// but what calls for special attention is that you will get blocked with the last
// Pool.Invoke() call once the current Pool runs out of its capacity, and to avoid this,
// you should instantiate a PoolWithFunc with fleets.WithNonblocking(true).
func (p *PoolWithFunc) Invoke(arg any) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.inputArg(arg)
	}
	return err
}

// NewPoolWithFunc instantiates a PoolWithFunc with customized options.
func NewPoolWithFunc(size int, pf func(any), options ...Option) (*PoolWithFunc, error) {
	if pf == nil {
		return nil, ErrLackPoolFunc
	}

	pc, err := newPool(size, options...)
	if err != nil {
		return nil, err
	}

	pool := &PoolWithFunc{
		poolCommon: pc,
		fn:         pf,
	}

	pool.workerCache.New = func() any {
		return &goWorkerWithFunc{
			pool: pool,
			arg:  make(chan any, workerChanCap),
		}
	}

	return pool, nil
}
