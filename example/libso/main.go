package main

import (
	"context"
	"fmt"

	"mlib.com/mrun/example/libso/test"
)

func Init(args ...interface{}) error {
	test.Init(args...)
	fmt.Println("init")
	return nil
}
func Destroy() {
	fmt.Println("Destroy")
}
func RunOnce(ctx context.Context) error {
	return nil
}

func UserData() interface{} {
	fmt.Println("UserData")
	return nil
}
