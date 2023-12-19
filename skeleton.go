package mrun

import (
	"context"
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

// Context interface contains an optional Context function which a Service can implement.
// When implemented the context.Done() channel will be used in addition to signal handling
// to exit a process.
type Context interface {
	Context() context.Context
}

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
	var ctx context.Context
	if s, ok := m.(Context); ok {
		ctx = s.Context()
	} else {
		ctx = context.Background()
	}
	mSkeleton.ctx, mSkeleton.ctxCancelFunc = context.WithCancel(ctx)

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

	select {
	case sigQuit := <-signalChan:
		log.Printf("[D]%s Server closing by signal %v\n", os.Args[0], sigQuit)
	case <-ctx.Done():
		log.Printf("[D]%s Server closing by context done\n", os.Args[0])
	}

	mSkeleton.Destroy()
	WorkerRelease()
	log.Printf("[D]%s Server End!\n", os.Args[0])
	return nil
}
