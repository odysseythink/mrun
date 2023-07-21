package mrun

import (
	"container/list"
	"fmt"
	"log"
)

type ParallelDataFlow struct {
	BaseDataFlow
}

func (df *ParallelDataFlow) Register(p IDataProcessor, options []DataFlowOption, args ...interface{}) error {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}

	if df.Contains(p) {
		log.Printf("[E]already register\n")
		return fmt.Errorf("already register")
	}

	info := &dataProcessorInfo{
		p:      p,
		exitCh: make(chan struct{}),
		order:  999,
	}
	if args != nil {
		info.args = make([]interface{}, 0)
		info.args = append(info.args, args...)
	}

	for _, v := range options {
		v(info)
	}
	if df.ctx != nil {
		err := p.Init(args...)
		if err != nil {
			log.Printf("[E]data processor init failed:%v\n", err)
			return fmt.Errorf("data processor init failed:%v", err)
		}
	}
	df.addDataProcessor(info)
	if df.ctx != nil {
		df.runDataProcessor(info)
	}
	return nil
}

func (df *ParallelDataFlow) Process(msg interface{}) (interface{}, error) {
	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	var outmsgs []interface{}
	e := df.processors.Front()
	for e != nil {
		outmsg, err := e.Value.(*dataProcessorInfo).p.Process(msg)
		if err != nil {
			log.Printf("[E]process failed:%v\n", err)
			if e.Value.(*dataProcessorInfo).processFaildFunc != nil {
				df.processorsMux.RUnlock()
				return e.Value.(*dataProcessorInfo).processFaildFunc(msg, err)
			} else {
				return nil, err
			}
		}
		if outmsgs == nil {
			outmsgs = make([]interface{}, 0)
		}
		outmsgs = append(outmsgs, outmsg)
		e = e.Next()
	}
	df.processorsMux.RUnlock()

	return outmsgs, nil
}
