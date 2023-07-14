package mrun

import (
	"context"
)

const (
	DEFAULT_MODULE_PRIORITY = 9999
)

type IModule interface {
	Init(args ...interface{}) error
	Destroy()
	RunOnce(ctx context.Context) error
	UserData() interface{}
}
