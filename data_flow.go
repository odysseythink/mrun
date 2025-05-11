package mrun

import (
	"context"
	"log"
)

type DataFlowOption func(*dataProcessorInfo)

func NewDataFlowProcessorErrorOption(cb func(IDataProcessor, error)) func(info *dataProcessorInfo) {
	return func(info *dataProcessorInfo) {
		if cb == nil {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.onProcessorError = cb
	}
}

func NewDataFlowProcessFaildOption(cb func(msg any, err error) (any, error)) func(info *dataProcessorInfo) {
	return func(info *dataProcessorInfo) {
		if cb == nil {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.processFaildFunc = cb
	}
}

func NewDataFlowOrderOption(order uint) func(info *dataProcessorInfo) {
	return func(info *dataProcessorInfo) {
		info.order = order
	}
}

type IDataProcessor interface {
	Init(args ...any) error
	RunOnce(ctx context.Context) error
	Destroy()
	UserData() any
	MsgCheck(msg any) error
	Process(msg any) (any, error)
}

type dataProcessorInfo struct {
	p                IDataProcessor
	args             []any
	exitCh           chan struct{}
	onProcessorError func(IDataProcessor, error)
	processFaildFunc func(msg any, err error) (any, error)
	order            uint
}

type IDataFlow interface {
	Contains(p IDataProcessor) bool
	GetDataProcessorInfo(p IDataProcessor) *dataProcessorInfo
	DeleteDataProcessorInfo(p IDataProcessor)
	Register(p IDataProcessor, options []DataFlowOption, args ...any) error
	UnRegister(p IDataProcessor) error
	Init() error
	ProcessorNum() int
	Destroy()
	Range(cb func(m IDataProcessor) bool)
	Process(msg any) (any, error)
}

func NewDataFlow(protocol string) IDataFlow {
	if protocol == "sequence" {
		return &SequenceDataFlow{}
	} else if protocol == "parallel" {
		return &ParallelDataFlow{}
	} else {
		log.Printf("[E]NewDataFlow failed:unsupported protocal(%s)\n", protocol)
		return nil
	}
}
