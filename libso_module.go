package mrun

import (
	"context"
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
