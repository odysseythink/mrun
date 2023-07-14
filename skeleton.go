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

func Run(m IModule) error {
	mSkeleton.Register(m, nil, nil)
	err := mSkeleton.Init()
	if err != nil {
		log.Printf("[E]skeleton init failed:%v\n", err)
		return err
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	mSkeleton.Destroy()
	WorkerRelease()
	log.Printf("[D]%s Server End!(closing by signal %v)\n", os.Args[0], sig)
	return nil
}
