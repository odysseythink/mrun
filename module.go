package mrun

import (
	"context"
)

const (
	DEFAULT_MODULE_PRIORITY = 9999
)

type IModule interface {
	Init(args ...any) error
	Destroy()
	RunOnce(ctx context.Context) error
	UserData() any
}
