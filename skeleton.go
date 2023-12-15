package mrun

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	mSkeleton = &ModuleMgr{
		name: "root_module_mgr",
	}
)

func Register(m IModule, options []ModuleMgrOption, args []interface{}) error {
	return mSkeleton.Register(m, options, args...)
}

func RegisterLibso(libname string, options []ModuleMgrOption, args []interface{}) error {
	return mSkeleton.RegisterLibso(libname, options, args...)
}

func RegisterLibsoWithModule(libname, modulename string, options []ModuleMgrOption, args []interface{}) error {
	return mSkeleton.RegisterLibsoWithModule(libname, modulename, options, args...)
}

func Run(m IModule, sig ...os.Signal) error {
	if m != nil {
		mSkeleton.Register(m, nil, nil)
	}
	err := mSkeleton.Init()
	if err != nil {
		log.Printf("[E]skeleton init failed:%v\n", err)
		return err
	}
	if len(sig) == 0 {
		sig = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, sig...)

	sigQuit := <-signalChan

	mSkeleton.Destroy()
	WorkerRelease()
	log.Printf("[D]%s Server End!(closing by signal %v)\n", os.Args[0], sigQuit)
	return nil
}
