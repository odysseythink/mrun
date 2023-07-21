package mrun

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type DataFlowOption func(*DataFlow, *dataProcessorInfo)

func NewDataFlowProcessorErrorOption(cb func(IDataProcessor, error)) func(df *DataFlow, info *dataProcessorInfo) {
	return func(df *DataFlow, info *dataProcessorInfo) {
		if cb == nil {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.onProcessorError = cb
	}
}

func NewDataFlowProcessFaildOption(cb func(msg interface{}, err error) (interface{}, error)) func(df *DataFlow, info *dataProcessorInfo) {
	return func(df *DataFlow, info *dataProcessorInfo) {
		if cb == nil {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.processFaildFunc = cb
	}
}

type IDataProcessor interface {
	Init(args ...interface{}) error
	RunOnce(ctx context.Context) error
	Destroy()
	UserData() interface{}
	Process(msg interface{}) (interface{}, error)
}

type dataProcessorInfo struct {
	p                IDataProcessor
	args             []interface{}
	exitCh           chan struct{}
	onProcessorError func(IDataProcessor, error)
	processFaildFunc func(msg interface{}, err error) (interface{}, error)
	order            uint
}

type DataFlow struct {
	processorsMux sync.RWMutex
	processors    *list.List
	ctx           context.Context
	wg            sync.WaitGroup
	ctxCancelFunc context.CancelFunc
	initOnce      sync.Once
}

func (df *DataFlow) Contains(p IDataProcessor) bool {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return false
	}

	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	e := df.processors.Front()
	for e != nil {
		if e.Value.(*dataProcessorInfo).p == p {
			df.processorsMux.RUnlock()
			return true
		}
		e = e.Next()
	}
	df.processorsMux.RUnlock()
	return false
}

func (df *DataFlow) GetDataProcessorInfo(p IDataProcessor) *dataProcessorInfo {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return nil
	}

	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	e := df.processors.Front()
	for e != nil {
		if e.Value.(*dataProcessorInfo).p == p {
			info := e.Value.(*dataProcessorInfo)
			df.processorsMux.RUnlock()
			return info
		}
		e = e.Next()
	}
	df.processorsMux.RUnlock()
	return nil
}

func (df *DataFlow) DeleteDataProcessorInfo(p IDataProcessor) {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return
	}

	df.processorsMux.Lock()
	if df.processors == nil {
		df.processors = list.New()
	}
	e := df.processors.Front()
	for e != nil {
		if e.Value.(*dataProcessorInfo).p == p {
			df.processors.Remove(e)
			df.processorsMux.RUnlock()
			return
		}
		e = e.Next()
	}
	df.processorsMux.Unlock()
}

func (df *DataFlow) addDataProcessor(info *dataProcessorInfo) {
	if info == nil {
		log.Printf("[E]invalid arg\n")
		return
	}
	df.processorsMux.Lock()
	e := df.processors.Front()
	for e != nil {
		if e.Value.(*dataProcessorInfo).order >= info.order {
			df.processors.InsertAfter(info, e)
			df.processorsMux.Unlock()
			return
		}
		e = e.Next()
	}
	df.processors.PushBack(info)
	df.processorsMux.Unlock()
}

func (df *DataFlow) Register(order uint, p IDataProcessor, options []DataFlowOption, args ...interface{}) error {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}

	if df.Contains(p) {
		log.Printf("[E]already register\n")
		return fmt.Errorf("already register")
	}
	if df.GetDataProcessorByOrder(order) != nil {
		log.Printf("[E]order(%d) already register\n", order)
		return fmt.Errorf("order(%d) already register", order)
	}
	if df.ctx != nil {
		err := p.Init(args...)
		if err != nil {
			log.Printf("[E]data processor init failed:%v\n", err)
			return fmt.Errorf("data processor init failed:%v", err)
		}
	}
	info := &dataProcessorInfo{
		p:      p,
		exitCh: make(chan struct{}),
		order:  order,
	}
	if args != nil {
		info.args = make([]interface{}, 0)
		info.args = append(info.args, args...)
	}

	for _, v := range options {
		v(df, info)
	}
	df.addDataProcessor(info)
	if df.ctx != nil {
		df.runDataProcessor(info)
	}
	return nil
}

func (df *DataFlow) UnRegister(p IDataProcessor) error {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}
	info := df.GetDataProcessorInfo(p)
	if info == nil {
		log.Printf("[E]data processor not register")
		return fmt.Errorf("data processor not register")
	}
	// df.wg.Add(1)
	// go func() {

	if df.ctx != nil {
		select {
		case info.exitCh <- struct{}{}:
			break
		default:
			// log.Println("notice exitCh failed")
			break
		}
	}
	info.p.Destroy()
	df.DeleteDataProcessorInfo(p)
	// 	df.wg.Done()
	// }()
	return nil
}

func (df *DataFlow) Init() error {
	var err error
	df.initOnce.Do(func() {
		df.processorsMux.RLock()
		if df.processors == nil {
			df.processors = list.New()
		}
		e := df.processors.Front()
		for e != nil {
			err = e.Value.(*dataProcessorInfo).p.Init(e.Value.(*dataProcessorInfo).args...)
			if err != nil {
				df.processorsMux.RUnlock()
				return
			}
			e = e.Next()
		}
		df.processorsMux.RUnlock()

		df.ctx, df.ctxCancelFunc = context.WithCancel(context.Background())
		df.processorsMux.RLock()
		e = df.processors.Front()
		for e != nil {
			df.runDataProcessor(e.Value.(*dataProcessorInfo))
			e = e.Next()
		}
		df.processorsMux.RUnlock()
	})

	return err
}

func (df *DataFlow) ModuleNum() int {
	df.processorsMux.RLock()
	num := df.processors.Len()
	df.processorsMux.RUnlock()
	return num
}

func (df *DataFlow) Destroy() {
	if df.ctxCancelFunc != nil {
		df.ctxCancelFunc()

		df.processorsMux.Lock()
		if df.processors == nil {
			df.processors = list.New()
		}
		e := df.processors.Front()
		for e != nil {
			df.destroy(e.Value.(*dataProcessorInfo))
			tmp := e.Next()
			df.processors.Remove(e)
			e = tmp

		}
		df.processorsMux.Unlock()
	}
	df.wg.Wait()
}

func (df *DataFlow) Range(cb func(m IDataProcessor) bool) {
	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	e := df.processors.Front()
	for e != nil {
		if !cb(e.Value.(*dataProcessorInfo).p) {
			df.processorsMux.RUnlock()
			return
		}
		e = e.Next()
	}
	df.processorsMux.RUnlock()
}

func (df *DataFlow) GetDataProcessorByOrder(order uint) IDataProcessor {
	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	e := df.processors.Front()
	for e != nil {
		if e.Value.(*dataProcessorInfo).order == order {
			df.processorsMux.RUnlock()
			return e.Value.(*dataProcessorInfo).p
		}
		e = e.Next()
	}
	df.processorsMux.RUnlock()
	return nil
}

func (df *DataFlow) destroy(info *dataProcessorInfo) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			log.Printf("[E]%v: %s\n", r, buf[:l])
		}
	}()
	if info != nil && info.p != nil {
		info.p.Destroy()
	}
}

func (df *DataFlow) runDataProcessor(info *dataProcessorInfo) {
	if info == nil || info.p == nil {
		return
	}

	df.wg.Add(1)
	go func() {
		var err error
		timer := time.NewTimer(1 * time.Millisecond)
	LOOP:
		for {
			select {
			case <-info.exitCh:
				log.Printf("[D]exit current data processor\n")
				break LOOP
			case <-df.ctx.Done():
				log.Printf("[D]context done\n")
				break LOOP
			case <-timer.C:
				err = info.p.RunOnce(df.ctx)
				if err != nil {
					log.Printf("[D]data processor RunOnce accur err(%v), exit data processor\n", err)
					df.UnRegister(info.p)
					if info.onProcessorError != nil {
						WorkerSubmit(func() {
							info.onProcessorError(info.p, err)
						})
					}
					break LOOP
				}
				timer.Reset(1 * time.Millisecond)
			}
		}
		df.wg.Done()
	}()
}

func (df *DataFlow) Process(msg interface{}) (interface{}, error) {
	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	var err error
	var inmsg interface{} = msg
	var outmsg interface{}
	e := df.processors.Front()
	for e != nil {
		outmsg, err = e.Value.(*dataProcessorInfo).p.Process(inmsg)
		if err != nil {
			log.Printf("[E]process failed:%v\n", err)
			if e.Value.(*dataProcessorInfo).processFaildFunc != nil {
				df.processorsMux.RUnlock()
				return e.Value.(*dataProcessorInfo).processFaildFunc(inmsg, err)
			} else {
				return nil, err
			}
		}

		inmsg = outmsg
		e = e.Next()
	}
	df.processorsMux.RUnlock()

	return outmsg, nil
}
