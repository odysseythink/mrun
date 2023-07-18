package main

import (
	"log"

	"mlib.com/mrun"
)

func main() {
	err := mrun.RegisterLibso("libso.so", nil, nil)
	if err != nil {
		log.Printf("mrun.RegisterLibso(\"libso.so\", nil, nil) failed:%v\n", err)
		return
	}
	err = mrun.RegisterLibsoWithModule("TestModule.so", "TestModule", nil, nil)
	if err != nil {
		log.Printf("mrun.RegisterLibso(\"TestModule\", nil, nil) failed:%v\n", err)
		return
	}
	mrun.Run(nil)
}
