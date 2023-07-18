package mrun

import (
	"context"
	"fmt"
	"log"
	"plugin"
	"strings"
)

type LibsoModule struct {
	init     func(args ...interface{}) error
	destroy  func()
	runOnce  func(ctx context.Context) error
	userData func() interface{}
}

func (m *LibsoModule) Init(args ...interface{}) error {
	if m.init != nil {
		return m.init()
	}
	return nil
}
func (m *LibsoModule) Destroy() {
	if m.destroy != nil {
		m.destroy()
	}
}
func (m *LibsoModule) RunOnce(ctx context.Context) error {
	if m.runOnce != nil {
		return m.runOnce(ctx)
	}
	return nil
}
func (m *LibsoModule) UserData() interface{} {
	if m.userData != nil {
		return m.userData()
	}
	return nil
}

func RegisterLibso(libname string, options []ModuleMgrOption, args []interface{}) error {
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
	m.init, ok = symbol.(func(args ...interface{}) error)
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
	return mSkeleton.Register(m, options, args...)
}

func RegisterLibsoWithModule(libname, modulename string, options []ModuleMgrOption, args []interface{}) error {
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

	return mSkeleton.Register(m, options, args...)
}
