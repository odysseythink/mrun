package mrun

import (
	"context"
	"fmt"
	"log"
	"plugin"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type ModuleMgrOption func(*ModuleMgr, IModule)

func NewPriorityModuleMgrOption(priority int) func(mgr *ModuleMgr, m IModule) {
	return func(mgr *ModuleMgr, m IModule) {
		if priority < 0 {
			log.Printf("[E]invalid arg\n")
			return
		}
		if val, ok := mgr.modules.Load(m); !ok {
			log.Printf("[E]module not saved")
			return
		} else {
			info := val.(*moduleInfo)
			mgr.preInitModsMux.Lock()
			if mgr.preInitMods == nil {
				mgr.preInitMods = make(map[int]map[*moduleInfo]struct{})
			}
			if mgr.preInitMods[priority] == nil {
				mgr.preInitMods[priority] = make(map[*moduleInfo]struct{})
			}
			if _, ok := mgr.preInitMods[priority][info]; ok {
				mgr.preInitModsMux.Unlock()
				log.Printf("[E]already exist\n")
				return
			}

			mgr.preInitMods[priority][info] = struct{}{}
			mgr.preInitModsMux.Unlock()
		}
	}
}

func NewModuleErrorOption(cb func(IModule, error)) func(mgr *ModuleMgr, m IModule) {
	return func(mgr *ModuleMgr, m IModule) {
		if cb == nil {
			log.Printf("[E]invalid arg\n")
			return
		}
		if val, ok := mgr.modules.Load(m); !ok {
			log.Printf("[E]module not saved")
			return
		} else {
			info := val.(*moduleInfo)
			info.onModuleError = cb
		}
	}
}

func NewModuleAliasOption(name string) func(mgr *ModuleMgr, m IModule) {
	return func(mgr *ModuleMgr, m IModule) {
		if name == "" {
			log.Printf("[E]invalid arg\n")
			return
		}
		if val, ok := mgr.modules.Load(m); !ok {
			log.Printf("[E]module not saved")
			return
		} else {
			info := val.(*moduleInfo)
			info.alias = name
		}
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
}

type ModuleMgr struct {
	name    string
	modules sync.Map

	preInitModsMux sync.Mutex
	preInitMods    map[int]map[*moduleInfo]struct{}
	ctx            context.Context
	wg             sync.WaitGroup
	ctxCancelFunc  context.CancelFunc
	initOnce       sync.Once
}

func (mgr *ModuleMgr) Register(m IModule, options []ModuleMgrOption, args ...interface{}) error {
	if m == nil {
		log.Printf("[E]invalid arg\n")
		return fmt.Errorf("invalid arg")
	}
	if _, ok := mgr.modules.Load(m); ok {
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
	}
	if args != nil {
		info.args = make([]interface{}, 0)
		info.args = append(info.args, args...)
	}

	mgr.modules.Store(m, info)
	for _, v := range options {
		v(mgr, m)
	}
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
	if _, ok := mgr.modules.Load(m); !ok {
		log.Printf("[E]module not register")
		return fmt.Errorf("module not register")
	}
	// mgr.wg.Add(1)
	// go func() {
	val, _ := mgr.modules.Load(m)
	if mgr.ctx != nil {
		select {
		case val.(*moduleInfo).exitCh <- struct{}{}:
			break
		default:
			// log.Println("notice exitCh failed")
			break
		}
	}
	val.(*moduleInfo).m.Destroy()
	mgr.modules.Delete(m)
	mgr.preInitModsMux.Lock()
	for k, v := range mgr.preInitMods {
		found := false
		for info := range v {
			if info.m == m {
				delete(mgr.preInitMods[k], info)
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	mgr.preInitModsMux.Unlock()
	// 	mgr.wg.Done()
	// }()
	return nil
}

func (mgr *ModuleMgr) Init() error {
	var err error
	mgr.initOnce.Do(func() {
		modInitFlagMap := make(map[*moduleInfo]bool)
		priorities := make([]int, 0)
		mgr.preInitModsMux.Lock()
		for k := range mgr.preInitMods {
			priorities = append(priorities, k)
		}
		sort.Ints(priorities)
		for _, v := range priorities {
			for info := range mgr.preInitMods[v] {
				if isInit, ok := modInitFlagMap[info]; !ok || !isInit {
					err = info.m.Init(info.args...)
					if err != nil {
						return
					}
					modInitFlagMap[info] = true
				}
			}
		}
		mgr.preInitMods = nil
		mgr.preInitModsMux.Unlock()

		mgr.modules.Range(func(key, value interface{}) bool {
			if _, ok := key.(IModule); !ok {
				log.Printf("[E]modules key not save IModule")
				err = fmt.Errorf("modules key not save IModule")
				return false
			} else {
				if info, ok := value.(*moduleInfo); !ok {
					log.Printf("[E]modules value not save moduleInfo pointer")
					err = fmt.Errorf("modules value not save moduleInfo pointer")
					return false
				} else {
					if isInit, ok := modInitFlagMap[info]; !ok || !isInit {
						err = info.m.Init(info.args...)
						if err != nil {
							return false
						}
						modInitFlagMap[info] = true
					}
				}
			}
			return true
		})
		if err != nil {
			return
		}
		mgr.ctx, mgr.ctxCancelFunc = context.WithCancel(context.Background())

		mgr.modules.Range(func(key, value interface{}) bool {
			mgr.runModule(value.(*moduleInfo))
			return true
		})
	})

	return err
}

func (mgr *ModuleMgr) ModuleNum() int {
	num := 0
	mgr.modules.Range(func(key, value interface{}) bool {
		num++
		return true
	})
	return num
}

func (mgr *ModuleMgr) Destroy() {
	if mgr.ctxCancelFunc != nil {
		mgr.ctxCancelFunc()
		mgr.modules.Range(func(key, value interface{}) bool {
			mgr.destroy(value.(*moduleInfo))
			mgr.modules.Delete(key)
			return true
		})
	}
	mgr.wg.Wait()
}

func (mgr *ModuleMgr) Range(cb func(m IModule) bool) {
	mgr.modules.Range(func(key, value interface{}) bool {
		return cb(value.(*moduleInfo).m)
	})
}

func (mgr *ModuleMgr) GetModulesByAlias(name string) []IModule {
	if name == "" {
		if name == "" {
			log.Printf("[E]invalid arg\n")
			return nil
		}
	}
	var ret []IModule
	mgr.modules.Range(func(key, value interface{}) bool {
		if info, ok := value.(*moduleInfo); ok && info != nil {
			if info.alias == name {
				if ret == nil {
					ret = make([]IModule, 0)
				}
				ret = append(ret, info.m)
			}
		}
		return true
	})
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
		timer := time.NewTimer(1 * time.Millisecond)
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
				timer.Reset(1 * time.Millisecond)
			}
		}
		mgr.wg.Done()
	}()
}
