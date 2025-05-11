package main

import (
	"context"
	"fmt"
)

// go build -buildmode=c-shared -o plugin.dll main.go

type testModule struct {
}

func (m *testModule) Init(args ...interface{}) error {
	fmt.Println("Init")
	return nil
}

func (m *testModule) RunOnce(ctx context.Context) error {
	return nil
}
func (m *testModule) Destroy() {
	fmt.Println("Destroy")
}
func (m *testModule) UserData() interface{} {
	return nil
}

var TestModule testModule

func main() {}
