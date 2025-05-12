package mrun

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type SignalInfo struct {
	Name       string
	Parameters []reflect.Type
	Callbacks  []reflect.Value
	FuncType   reflect.Type
}

type CallInfo struct {
	Args []reflect.Value
	Func reflect.Value
}

type Signal struct {
	name           string
	infos          map[string]*SignalInfo
	infosLock      sync.RWMutex
	callbackCh     chan *CallInfo
	concurrencyNum int
	subSigs        []*Signal
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

func (s *Signal) SendDirect(name string, args ...any) error {
	if name == "" {
		log.Printf("[E]missing signal name\n")
		return errors.New("missing signal name")
	}

	var firstletter rune = rune(name[0])
	if !unicode.IsLetter(firstletter) {
		log.Printf("[E]signal(%s) must start with letter\n", name)
		return fmt.Errorf("signal(%s) must start with letter", name)
	}
	name = strings.ToUpper(string(name[0])) + name[1:]
	s.infosLock.RLock()
	var info *SignalInfo
	if _, ok := s.infos[name]; !ok {
		log.Printf("[E]signal(%s) not exist\n", name)
		s.infosLock.RUnlock()
		return fmt.Errorf("signal(%s) not exist", name)
	}
	s.infosLock.RUnlock()
	info = s.infos[name]
	if len(info.Parameters) != len(args) {
		log.Printf("[E]argument %d length doesn't equal to provide length %d \n", len(info.Parameters), len(args))
		return fmt.Errorf("argument %d length doesn't equal to provide length %d ", len(info.Parameters), len(args))
	}
	argValues := make([]reflect.Value, 0, len(info.Parameters))
	for idx, v := range info.Parameters {
		t := reflect.TypeOf(args[idx])
		if v != t {
			log.Printf("[E]type(argument[%d])=%s doesn't match the type of %s \n", idx, info.Parameters[idx].Name(), t.Name())
			return fmt.Errorf("[E]type(argument[%d])=%s doesn't match the type of %s", idx, info.Parameters[idx].Name(), t.Name())
		}
		argValues = append(argValues, reflect.ValueOf(args[idx]))
	}
	for _, cb := range info.Callbacks {
		cb.Call(argValues)
	}
	return nil
}

func (s *Signal) Send(name string, args ...any) error {
	if name == "" {
		log.Printf("[E]missing signal name\n")
		return errors.New("missing signal name")
	}

	var firstletter rune = rune(name[0])
	if !unicode.IsLetter(firstletter) {
		log.Printf("[E]signal(%s) must start with letter\n", name)
		return fmt.Errorf("signal(%s) must start with letter", name)
	}
	name = strings.ToUpper(string(name[0])) + name[1:]
	s.infosLock.RLock()
	var info *SignalInfo
	if _, ok := s.infos[name]; !ok {
		log.Printf("[E]signal(%s) not exist\n", name)
		s.infosLock.RUnlock()
		return fmt.Errorf("signal(%s) not exist", name)
	}
	s.infosLock.RUnlock()
	info = s.infos[name]
	if len(info.Parameters) != len(args) {
		log.Printf("[E]argument %d length doesn't equal to provide length %d \n", len(info.Parameters), len(args))
		return fmt.Errorf("argument %d length doesn't equal to provide length %d ", len(info.Parameters), len(args))
	}
	argValues := make([]reflect.Value, 0, len(info.Parameters))
	for idx, v := range info.Parameters {
		t := reflect.TypeOf(args[idx])
		if v != t {
			log.Printf("[E]type(argument[%d])=%s doesn't match the type of %s \n", idx, info.Parameters[idx].Name(), t.Name())
			return fmt.Errorf("[E]type(argument[%d])=%s doesn't match the type of %s", idx, info.Parameters[idx].Name(), t.Name())
		}
		argValues = append(argValues, reflect.ValueOf(args[idx]))
	}
	for _, cb := range info.Callbacks {
		s.callbackCh <- &CallInfo{
			Func: cb,
			Args: argValues,
		}
	}
	return nil
}

func (s *Signal) RegisterSigfunc(name string, sigfunc any) error {
	if name == "" {
		log.Printf("[E]missing signal name\n")
		return errors.New("missing signal name")
	}
	if sigfunc == nil {
		log.Printf("[E]missing signal func\n")
		return errors.New("missing signal func")
	}
	var firstletter rune = rune(name[0])
	if !unicode.IsLetter(firstletter) {
		log.Printf("[E]signal(%s) must start with letter\n", name)
		return fmt.Errorf("signal(%s) must start with letter", name)
	}
	name = strings.ToUpper(string(name[0])) + name[1:]
	if s.infos == nil {
		s.infos = make(map[string]*SignalInfo)
	}
	s.infosLock.RLock()
	if _, ok := s.infos[name]; ok {
		log.Printf("[E]signal(%s) already exist\n", name)
		s.infosLock.RUnlock()
		return fmt.Errorf("signal(%s) already exist", name)
	}
	s.infosLock.RUnlock()

	t := reflect.TypeOf(sigfunc)
	if t.Kind() != reflect.Func {
		log.Printf("[E]slot is not function\n")
		return errors.New("slot is not function")
	}
	info := &SignalInfo{
		Name:       name,
		FuncType:   t,
		Parameters: make([]reflect.Type, 0),
		Callbacks:  make([]reflect.Value, 0),
	}

	info.Parameters = make([]reflect.Type, 0, t.NumIn())
	for i := range t.NumIn() {
		arg := t.In(i)
		log.Printf("[D]argument %d is %s[%s] type \n", i, arg.Kind(), arg.Name())
		info.Parameters = append(info.Parameters, arg)
	}
	info.Callbacks = make([]reflect.Value, 0)
	s.infosLock.Lock()
	s.infos[name] = info
	s.infosLock.Unlock()
	return nil
}

func NewSignalFuncOption(name string, sigfunc any) func(*Signal) error {
	return func(s *Signal) error {
		if name == "" {
			log.Printf("[E]missing signal func name\n")
			return errors.New("missing signal func name")
		}
		if sigfunc == nil {
			log.Printf("[E]missing signal func\n")
			return errors.New("missing signal func")
		}

		return s.RegisterSigfunc(name, sigfunc)
	}
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

func NewSignal(name string, options ...SignalOption) (*Signal, error) {
	if name == "" {
		log.Printf("[E]missing name\n")
		return nil, errors.New("missing name")
	}
	mods := mSkeleton.GetModulesByAlias(name)
	if len(mods) > 0 {
		log.Printf("[E]signal(%s) already exist\n", name)
		return nil, fmt.Errorf("signal(%s) already exist", name)
	}
	s := &Signal{
		name:  name,
		infos: make(map[string]*SignalInfo),
	}
	for _, option := range options {
		err := option(s)
		if err != nil {
			log.Printf("[E]run signal option func failed:%v\n", err)
			return nil, err
		}
	}
	if len(s.infos) == 0 {
		log.Printf("[E]you must define at least one signal function\n")
		return nil, errors.New("you must define at least one signal function")
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
		s.subSigs = make([]*Signal, 0)
		for iLoop := range s.concurrencyNum {
			subs := &Signal{
				name:       name + strconv.Itoa(iLoop),
				callbackCh: s.callbackCh,
			}
			err := mSkeleton.Register(subs, []ModuleMgrOption{NewModuleAliasOption(subs.name)}, nil)
			if err != nil {
				mSkeleton.UnRegister(s)
				for _, v := range s.subSigs {
					mSkeleton.UnRegister(v)
				}
				log.Printf("[E]register signal model failed:%v\n", err)
				return nil, err
			}
			s.subSigs = append(s.subSigs, subs)
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
	receiver.infosLock.Lock()
	for k, v := range receiver.infos {
		if v.FuncType == t {
			receiver.infos[k].Callbacks = append(receiver.infos[k].Callbacks, value)
			receiver.infosLock.Unlock()
			return nil
		}
	}
	receiver.infosLock.Unlock()

	log.Printf("[E]slot not exist signal\n")
	return errors.New("slot not exist signal")
}
