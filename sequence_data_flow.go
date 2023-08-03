package mrun

import (
	"container/list"
	"fmt"
	"log"
)

type SequenceDataFlow struct {
	BaseDataFlow
}

func (df *SequenceDataFlow) Register(p IDataProcessor, options []DataFlowOption, args ...interface{}) error {
	if p == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}

	if df.Contains(p) {
		log.Printf("[E]already register\n")
		return fmt.Errorf("already register")
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
		order:  999,
	}
	if args != nil {
		info.args = make([]interface{}, 0)
		info.args = append(info.args, args...)
	}

	for _, v := range options {
		v(info)
	}
	if df.getDataProcessorByOrder(info.order) != nil {
		log.Printf("[E]order(%d) already register\n", info.order)
		return fmt.Errorf("order(%d) already register", info.order)
	}
	df.addDataProcessor(info)
	if df.ctx != nil {
		df.runDataProcessor(info)
	}
	return nil
}

func (df *SequenceDataFlow) getDataProcessorByOrder(order uint) IDataProcessor {
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

func (df *SequenceDataFlow) Process(msg interface{}) (interface{}, error) {
	df.processorsMux.RLock()
	if df.processors == nil {
		df.processors = list.New()
	}
	var err error
	var inmsg interface{} = msg
	var outmsg interface{}
	e := df.processors.Front()
	for e != nil {
		err = e.Value.(*dataProcessorInfo).p.MsgCheck(msg)
		if err != nil {
			log.Printf("[E]MsgCheck failed:%v\n", err)
			if e.Value.(*dataProcessorInfo).processFaildFunc != nil {
				df.processorsMux.RUnlock()
				return e.Value.(*dataProcessorInfo).processFaildFunc(msg, err)
			} else {
				return nil, err
			}
		}
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
