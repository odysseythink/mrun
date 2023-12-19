package mrun

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"plugin"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ModuleMgrOption func(*ModuleMgr, *moduleInfo)

func NewPriorityModuleMgrOption(priority int) func(mgr *ModuleMgr, info *moduleInfo) {
	return func(mgr *ModuleMgr, info *moduleInfo) {
		if priority < 0 {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.order = uint(priority)
	}
}

func NewModuleErrorOption(cb func(IModule, error)) func(mgr *ModuleMgr, info *moduleInfo) {
	return func(mgr *ModuleMgr, info *moduleInfo) {
		if cb == nil {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.onModuleError = cb
	}
}

func NewModuleAliasOption(name string) func(mgr *ModuleMgr, info *moduleInfo) {
	return func(mgr *ModuleMgr, info *moduleInfo) {
		if name == "" {
			log.Printf("[E]invalid arg\n")
			return
		}
		info.alias = name
	}
}

func NewModuleRunPeriodOption(msec uint) func(mgr *ModuleMgr, info *moduleInfo) {
	return func(mgr *ModuleMgr, info *moduleInfo) {
		if msec == 0 {
			msec = 1
		}
		info.period = msec
	}
}

func NewModuleMgr(name string) *ModuleMgr {
	if name == "" {
		log.Printf("[E]invalid arg\n")
		return nil
	}
	return &ModuleMgr{name: name}
}

type moduleInfo struct {
	m             IModule
	args          []interface{}
	exitCh        chan struct{}
	onModuleError func(IModule, error)
	alias         string
	order         uint
	period        uint
}

type ModuleMgr struct {
	name          string
	modulesMux    sync.RWMutex
	modules       *list.List
	ctx           context.Context
	wg            sync.WaitGroup
	ctxCancelFunc context.CancelFunc
	initOnce      sync.Once
}

func (mgr *ModuleMgr) Contains(m IModule) bool {
	if m == nil {
		log.Printf("[E]invalid arg\n")
		return false
	}

	mgr.modulesMux.RLock()
	if mgr.modules == nil {
		mgr.modules = list.New()
	}
	e := mgr.modules.Front()
	for e != nil {
		if e.Value.(*moduleInfo).m == m {
			mgr.modulesMux.RUnlock()
			return true
		}
		e = e.Next()
	}
	mgr.modulesMux.RUnlock()
	return false
}

func (mgr *ModuleMgr) GetModuleInfo(m IModule) *moduleInfo {
	if m == nil {
		log.Printf("[E]invalid arg\n")
		return nil
	}

	mgr.modulesMux.RLock()
	if mgr.modules == nil {
		mgr.modules = list.New()
	}
	e := mgr.modules.Front()
	for e != nil {
		if e.Value.(*moduleInfo).m == m {
			info := e.Value.(*moduleInfo)
			mgr.modulesMux.RUnlock()
			return info
		}
		e = e.Next()
	}
	mgr.modulesMux.RUnlock()
	return nil
}

func (mgr *ModuleMgr) DeleteModuleInfo(m IModule) {
	if m == nil {
		log.Printf("[E]invalid arg\n")
		return
	}

	mgr.modulesMux.Lock()
	if mgr.modules == nil {
		mgr.modules = list.New()
	}
	e := mgr.modules.Front()
	for e != nil {
		if e.Value.(*moduleInfo).m == m {
			mgr.modules.Remove(e)
			mgr.modulesMux.RUnlock()
			return
		}
		e = e.Next()
	}
	mgr.modulesMux.Unlock()
}

func (mgr *ModuleMgr) addModule(info *moduleInfo) {
	if info == nil {
		log.Printf("[E]invalid arg\n")
		return
	}
	mgr.modulesMux.Lock()
	e := mgr.modules.Front()
	for e != nil {
		if e.Value.(*moduleInfo).order >= info.order {
			mgr.modules.InsertAfter(info, e)
			mgr.modulesMux.Unlock()
			return
		}
		e = e.Next()
	}
	mgr.modules.PushBack(info)
	mgr.modulesMux.Unlock()
}

func (mgr *ModuleMgr) Register(m IModule, options []ModuleMgrOption, args ...interface{}) error {
	if m == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}

	if mgr.Contains(m) {
		log.Printf("[E]already register")
		return fmt.Errorf("already register")
	}
	if mgr.ctx != nil {
		err := m.Init(args...)
		if err != nil {
			log.Printf("[E]module init failed:%v\n", err)
			return fmt.Errorf("module init failed:%v", err)
		}
	}
	info := &moduleInfo{
		m:      m,
		exitCh: make(chan struct{}),
		order:  9999,
		period: 1,
	}
	if args != nil {
		info.args = make([]interface{}, 0)
		info.args = append(info.args, args...)
	}

	for _, v := range options {
		v(mgr, info)
	}
	mgr.addModule(info)
	if mgr.ctx != nil {
		mgr.runModule(info)
	}
	return nil
}

func (mgr *ModuleMgr) RegisterLibso(libname string, options []ModuleMgrOption, args ...interface{}) error {
	if !strings.HasSuffix(libname, ".so") {
		log.Printf("[E]libname(%s) must be a so lib\n", libname)
		return fmt.Errorf("libname(%s) must be a so lib", libname)
	}
	log.Printf("[D]ModuleName: %s\n", libname)
	// return mSkeleton.Register(m, options, args...)
	m := new(LibsoModule)
	plug, err := plugin.Open(libname)
	if err != nil {
		log.Printf("[E]load Module(%s) failed:%v\n", libname, err)
		return fmt.Errorf("load Module(%s) failed:%v", libname, err)
	}
	var symbol plugin.Symbol
	var ok bool
	symbol, err = plug.Lookup("Init")
	if err != nil {
		log.Printf("[E]Module(%s) Lookup Init function failed:%v\n", libname, err)
		return fmt.Errorf("Module(%s) Lookup Init function failed:%v", libname, err)
	}
	m.init, ok = symbol.(func(...interface{}) error)
	if !ok {
		log.Printf("[E]Module(%s) Init function has wrong type\n", libname)
		return fmt.Errorf("Module(%s) Init function has wrong type", libname)
	}

	symbol, err = plug.Lookup("Destroy")
	if err != nil {
		log.Printf("[E]Module(%s) Lookup Destroy function failed:%v\n", libname, err)
		return fmt.Errorf("Module(%s) Lookup Destroy function failed:%v", libname, err)
	}
	m.destroy, ok = symbol.(func())
	if !ok {
		log.Printf("[E]Module(%s) Destroy function has wrong type\n", libname)
		return fmt.Errorf("Module(%s) Destroy function has wrong type", libname)
	}

	symbol, err = plug.Lookup("RunOnce")
	if err != nil {
		log.Printf("[E]Module(%s) Lookup RunOnce function failed:%v", libname, err)
		return fmt.Errorf("Module(%s) Lookup RunOnce function failed:%v", libname, err)
	}
	m.runOnce, ok = symbol.(func(ctx context.Context) error)
	if !ok {
		log.Printf("[E]Module(%s) RunOnce function has wrong type", libname)
		return fmt.Errorf("Module(%s) RunOnce function has wrong type", libname)
	}

	symbol, err = plug.Lookup("UserData")
	if err != nil {
		log.Printf("[E]Module(%s) Lookup UserData function failed:%v", libname, err)
		return fmt.Errorf("Module(%s) Lookup UserData function failed:%v", libname, err)
	}
	m.userData, ok = symbol.(func() interface{})
	if !ok {
		log.Printf("[E]Module(%s) UserData function has wrong type", libname)
		return fmt.Errorf("Module(%s) UserData function has wrong type", libname)
	}
	if m.init == nil || m.destroy == nil || m.runOnce == nil || m.userData == nil {
		log.Printf("[E]Module(%s) both init, destroy, runOnce and userData must be provied\n", libname)
		return fmt.Errorf("Module(%s) both init, destroy, runOnce and userData must be provied", libname)
	}
	return mgr.Register(m, options, args...)
}

func (mgr *ModuleMgr) RegisterLibsoWithModule(libname, modulename string, options []ModuleMgrOption, args ...interface{}) error {
	if !strings.HasSuffix(libname, ".so") {
		log.Printf("[E]libname(%s) must be a so lib\n", libname)
		return fmt.Errorf("libname(%s) must be a so lib", libname)
	}
	log.Printf("[D]ModuleName: %s\n", modulename)
	// return mSkeleton.Register(m, options, args...)
	plug, err := plugin.Open(libname)
	if err != nil {
		log.Printf("[E]load Module(%s) failed:%v\n", libname, err)
		return fmt.Errorf("load Module(%s) failed:%v", libname, err)
	}
	var symbol plugin.Symbol
	symbol, err = plug.Lookup(modulename)
	if err != nil {
		log.Printf("[E]Module(%s) Lookup %s failed:%v\n", libname, modulename, err)
		return fmt.Errorf("Module(%s) Lookup %s failed:%v", libname, modulename, err)
	}
	fmt.Printf("-------symbol=%#v\n", symbol)
	m, ok := symbol.(IModule)
	if !ok {
		log.Printf("[E]Module(%s) must implement of IModule\n", modulename)
		return fmt.Errorf("Module(%s) must implement of IModule", modulename)
	}

	return mgr.Register(m, options, args...)
}

func (mgr *ModuleMgr) UnRegister(m IModule) error {
	if m == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}
	info := mgr.GetModuleInfo(m)
	if info == nil {
		log.Printf("[E]module not register")
		return fmt.Errorf("module not register")
	}
	// mgr.wg.Add(1)
	// go func() {

	if mgr.ctx != nil {
		select {
		case info.exitCh <- struct{}{}:
			break
		default:
			// log.Println("notice exitCh failed")
			break
		}
	}
	info.m.Destroy()
	mgr.DeleteModuleInfo(m)
	// 	mgr.wg.Done()
	// }()
	return nil
}

func (mgr *ModuleMgr) Init() error {
	var err error
	mgr.initOnce.Do(func() {
		mgr.modulesMux.RLock()
		if mgr.modules == nil {
			mgr.modules = list.New()
		}
		e := mgr.modules.Front()
		for e != nil {
			err = e.Value.(*moduleInfo).m.Init(e.Value.(*moduleInfo).args...)
			if err != nil {
				mgr.modulesMux.RUnlock()
				return
			}
			e = e.Next()
		}
		mgr.modulesMux.RUnlock()

		if mgr.ctx == nil {
			mgr.ctx, mgr.ctxCancelFunc = context.WithCancel(context.Background())
		}
		mgr.modulesMux.RLock()
		e = mgr.modules.Front()
		for e != nil {
			mgr.runModule(e.Value.(*moduleInfo))
			e = e.Next()
		}
		mgr.modulesMux.RUnlock()
	})

	return err
}

func (mgr *ModuleMgr) ModuleNum() int {
	mgr.modulesMux.RLock()
	num := mgr.modules.Len()
	mgr.modulesMux.RUnlock()
	return num
}

func (mgr *ModuleMgr) Destroy() {
	if mgr.ctxCancelFunc != nil {
		mgr.ctxCancelFunc()

		mgr.modulesMux.Lock()
		if mgr.modules == nil {
			mgr.modules = list.New()
		}
		e := mgr.modules.Front()
		for e != nil {
			mgr.destroy(e.Value.(*moduleInfo))
			tmp := e.Next()
			mgr.modules.Remove(e)
			e = tmp

		}
		mgr.modulesMux.Unlock()
	}
	mgr.wg.Wait()
}

func (mgr *ModuleMgr) Range(cb func(m IModule) bool) {
	mgr.modulesMux.RLock()
	if mgr.modules == nil {
		mgr.modules = list.New()
	}
	e := mgr.modules.Front()
	for e != nil {
		if !cb(e.Value.(*moduleInfo).m) {
			mgr.modulesMux.RUnlock()
			return
		}
		e = e.Next()
	}
	mgr.modulesMux.RUnlock()
}

func (mgr *ModuleMgr) GetModulesByAlias(name string) []IModule {
	if name == "" {
		if name == "" {
			log.Printf("[E]invalid arg\n")
			return nil
		}
	}
	var ret []IModule
	mgr.modulesMux.RLock()
	if mgr.modules == nil {
		mgr.modules = list.New()
	}
	e := mgr.modules.Front()
	for e != nil {
		if e.Value.(*moduleInfo).alias == name {
			if ret == nil {
				ret = make([]IModule, 0)
			}
			ret = append(ret, e.Value.(*moduleInfo).m)
		}
		e = e.Next()
	}
	mgr.modulesMux.RUnlock()
	return ret
}

func (mgr *ModuleMgr) destroy(info *moduleInfo) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			log.Printf("[E]%v: %s\n", r, buf[:l])
		}
	}()
	if info != nil && info.m != nil {
		info.m.Destroy()
	}
}

func (mgr *ModuleMgr) runModule(info *moduleInfo) {
	if info == nil || info.m == nil {
		return
	}

	mgr.wg.Add(1)
	go func() {
		var err error
		timer := time.NewTimer(time.Duration(info.period) * time.Millisecond)
	LOOP:
		for {
			select {
			case <-info.exitCh:
				log.Printf("[D]exit current module\n")
				break LOOP
			case <-mgr.ctx.Done():
				log.Printf("[D]context done\n")
				break LOOP
			case <-timer.C:
				err = info.m.RunOnce(mgr.ctx)
				if err != nil {
					log.Printf("[D]module RunOnce accur err(%v), exit module\n", err)
					mgr.UnRegister(info.m)
					if info.onModuleError != nil {
						WorkerSubmit(func() {
							info.onModuleError(info.m, err)
						})
					}
					break LOOP
				}
				timer.Reset(time.Duration(info.period) * time.Millisecond)
			}
		}
		mgr.wg.Done()
	}()
}
