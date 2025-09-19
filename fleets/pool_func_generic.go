package fleets

// PoolWithFuncGeneric is the generic version of PoolWithFunc.
type PoolWithFuncGeneric[T any] struct {
	*poolCommon

	// fn is the unified function for processing tasks.
	fn func(T)
}

// Invoke passes the argument to the pool to start a new task.
func (p *PoolWithFuncGeneric[T]) Invoke(arg T) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.(*goWorkerWithFuncGeneric[T]).arg <- arg
	}
	return err
}

// NewPoolWithFuncGeneric instantiates a PoolWithFuncGeneric[T] with customized options.
func NewPoolWithFuncGeneric[T any](size int, pf func(T), options ...Option) (*PoolWithFuncGeneric[T], error) {
	if pf == nil {
		return nil, ErrLackPoolFunc
	}

	pc, err := newPool(size, options...)
	if err != nil {
		return nil, err
	}

	pool := &PoolWithFuncGeneric[T]{
		poolCommon: pc,
		fn:         pf,
	}

	pool.workerCache.New = func() any {
		return &goWorkerWithFuncGeneric[T]{
			pool: pool,
			arg:  make(chan T, workerChanCap),
			exit: make(chan struct{}, 1),
		}
	}

	return pool, nil
}
