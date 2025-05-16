package mrun

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
)

type CallInfo struct {
	Args []reflect.Value
	Func reflect.Value
}

type Signal struct {
	name        string
	sigFuncType reflect.Type
	Parameters  []reflect.Type
	Callbacks   []reflect.Value
	sync.RWMutex
	callbackCh     chan *CallInfo
	concurrencyNum int
	sigConsumers   []*Signal
}

func (s *Signal) Init(args ...any) error {
	return nil
}
func (s *Signal) Destroy() {

}
func (s *Signal) RunOnce(ctx context.Context) error {
	select {
	case cinfo := <-s.callbackCh:
		if cinfo != nil {

			cinfo.Func.Call(cinfo.Args)
		}
	default:

	}
	return nil
}
func (s *Signal) UserData() any {
	return nil
}

func (s *Signal) EmitDirect(args ...any) error {
	if len(s.Parameters) != len(args) {
		log.Printf("[E]argument %d length doesn't equal to provide length %d \n", len(s.Parameters), len(args))
		return fmt.Errorf("argument %d length doesn't equal to provide length %d ", len(s.Parameters), len(args))
	}
	argValues := make([]reflect.Value, 0, len(s.Parameters))
	for idx, v := range s.Parameters {
		t := reflect.TypeOf(args[idx])
		if v != t {
			log.Printf("[E]type(argument[%d])=%s doesn't match the type of %s \n", idx, s.Parameters[idx].Name(), t.Name())
			return fmt.Errorf("[E]type(argument[%d])=%s doesn't match the type of %s", idx, s.Parameters[idx].Name(), t.Name())
		}
		argValues = append(argValues, reflect.ValueOf(args[idx]))
	}
	for _, cb := range s.Callbacks {
		cb.Call(argValues)
	}
	return nil
}

func (s *Signal) Emit(args ...any) error {
	if len(s.Parameters) != len(args) {
		log.Printf("[E]argument %d length doesn't equal to provide length %d \n", len(s.Parameters), len(args))
		return fmt.Errorf("argument %d length doesn't equal to provide length %d ", len(s.Parameters), len(args))
	}
	argValues := make([]reflect.Value, 0, len(s.Parameters))
	for idx, v := range s.Parameters {
		t := reflect.TypeOf(args[idx])
		if v != t {
			log.Printf("[E]type(argument[%d])=%s doesn't match the type of %s \n", idx, s.Parameters[idx].Name(), t.Name())
			return fmt.Errorf("[E]type(argument[%d])=%s doesn't match the type of %s", idx, s.Parameters[idx].Name(), t.Name())
		}
		argValues = append(argValues, reflect.ValueOf(args[idx]))
	}
	for _, cb := range s.Callbacks {
		s.callbackCh <- &CallInfo{
			Func: cb,
			Args: argValues,
		}
	}
	return nil
}

func NewSignalChCapOption(cap int) func(*Signal) error {
	return func(s *Signal) error {
		if cap <= 0 {
			log.Printf("[E]invalid signal ch capacity lenght\n")
			return errors.New("invalid signal ch capacity lenght")
		}
		s.callbackCh = make(chan *CallInfo, cap)

		return nil
	}
}

func NewSignalConcurrencyOption(num int) func(*Signal) error {
	return func(s *Signal) error {
		if num < 0 {
			log.Printf("[E]invalid signal concurrency num\n")
			return errors.New("invalid signal concurrency num")
		}
		s.concurrencyNum = num

		return nil
	}
}

type SignalOption func(*Signal) error

func NewSignal(name string, sigfunc any, options ...SignalOption) (*Signal, error) {
	if name == "" {
		log.Printf("[E]missing name\n")
		return nil, errors.New("missing name")
	}
	if sigfunc == nil {
		log.Printf("[E]missing signal func\n")
		return nil, errors.New("missing signal func")
	}
	mods := mSkeleton.GetModulesByAlias(name)
	if len(mods) > 0 {
		log.Printf("[E]signal(%s) already exist\n", name)
		return nil, fmt.Errorf("signal(%s) already exist", name)
	}
	s := &Signal{
		name:       name,
		Parameters: make([]reflect.Type, 0),
		Callbacks:  make([]reflect.Value, 0),
	}
	if name == "" {
		log.Printf("[E]missing signal func name\n")
		return nil, errors.New("missing signal func name")
	}

	t := reflect.TypeOf(sigfunc)
	if t.Kind() != reflect.Func {
		log.Printf("[E]sigfunc is not function\n")
		return nil, errors.New("sigfunc is not function")
	}
	s.sigFuncType = t

	for i := range t.NumIn() {
		arg := t.In(i)
		log.Printf("[D]argument %d is %s[%s] type \n", i, arg.Kind(), arg.Name())
		s.Parameters = append(s.Parameters, arg)
	}

	for _, option := range options {
		err := option(s)
		if err != nil {
			log.Printf("[E]run signal option func failed:%v\n", err)
			return nil, err
		}
	}

	if s.callbackCh == nil {
		s.callbackCh = make(chan *CallInfo, 1024)
	}
	err := mSkeleton.Register(s, []ModuleMgrOption{NewModuleAliasOption(name)}, nil)
	if err != nil {
		log.Printf("[E]register signal model failed:%v\n", err)
		return nil, err
	}
	if s.concurrencyNum > 0 {
		s.sigConsumers = make([]*Signal, 0)
		for iLoop := range s.concurrencyNum {
			subs := &Signal{
				name:       name + strconv.Itoa(iLoop),
				callbackCh: s.callbackCh,
			}
			err := mSkeleton.Register(subs, []ModuleMgrOption{NewModuleAliasOption(subs.name)}, nil)
			if err != nil {
				mSkeleton.UnRegister(s)
				for _, v := range s.sigConsumers {
					mSkeleton.UnRegister(v)
				}
				log.Printf("[E]register signal model failed:%v\n", err)
				return nil, err
			}
			s.sigConsumers = append(s.sigConsumers, subs)
		}
	}
	return s, nil
}

func Connect(receiver *Signal, slot any) error {
	if receiver == nil {
		log.Printf("[E]missing receiver\n")
		return errors.New("missing receiver")
	}
	value := reflect.ValueOf(slot)
	if value.Kind() != reflect.Func {
		log.Printf("[E]slot is not function\n")
		return errors.New("slot is not function")
	}
	t := reflect.TypeOf(slot)
	if receiver.sigFuncType != t {
		log.Printf("[E]slot (%s) is not matched signal(%s)\n", t.Name(), receiver.sigFuncType.Name())
		return errors.New("slot is not matched signal")
	}

	receiver.Lock()
	receiver.Callbacks = append(receiver.Callbacks, value)
	receiver.Unlock()

	return nil
}
