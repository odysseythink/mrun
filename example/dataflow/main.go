package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"mlib.com/mrun"
)

type Processor1 struct {
}

func (p Processor1) Init(args ...interface{}) error {
	fmt.Println("Processor1 Init")
	return nil
}
func (p Processor1) RunOnce(ctx context.Context) error {
	return nil
}
func (p Processor1) Destroy() {
	fmt.Println("Processor1 Destroy")
}
func (p Processor1) UserData() interface{} {
	return nil
}
func (p Processor1) Process(msg interface{}) (interface{}, error) {
	if msg == nil {
		fmt.Println("Processor1 Process failed:invalid arg")
		return nil, errors.New("Processor1 Process failed:invalid arg")
	}
	if val, ok := msg.(int); !ok {
		fmt.Println("Processor1 Process failed:invalid msg type")
		return nil, errors.New("Processor1 Process failed:invalid msg type")
	} else {
		fmt.Println("Processor1 Process success")
		return val + 1, nil
	}
}

type Processor2 struct {
}

func (p Processor2) Init(args ...interface{}) error {
	fmt.Println("Processor2 Init")
	return nil
}
func (p Processor2) RunOnce(ctx context.Context) error {
	return nil
}
func (p Processor2) Destroy() {
	fmt.Println("Processor2 Destroy")
}
func (p Processor2) UserData() interface{} {
	return p
}
func (p Processor2) Process(msg interface{}) (interface{}, error) {
	if msg == nil {
		fmt.Println("Processor2 Process failed:invalid arg")
		return nil, errors.New("Processor2 Process failed:invalid arg")
	}
	if val, ok := msg.(int); !ok {
		fmt.Println("Processor2 Process failed:invalid msg type")
		return nil, errors.New("Processor2 Process failed:invalid msg type")
	} else {
		fmt.Println("Processor2 Process success")
		return val - 1, nil
	}
}

func main() {
	var df mrun.DataFlow
	err := df.Register(0, &Processor1{}, nil, nil)
	if err != nil {
		fmt.Println("DataFlow Register Processor1 failed:", err)
		return
	}
	err = df.Register(1, &Processor2{}, nil, nil)
	if err != nil {
		fmt.Println("DataFlow Register Processor2 failed:", err)
		return
	}
	err = df.Init()
	if err != nil {
		fmt.Println("DataFlow init failed:", err)
		return
	}
	var wg sync.WaitGroup
	for iLoop := 0; iLoop < 100; iLoop++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ret, err := df.Process(idx)
			if err != nil {
				fmt.Printf("DataFlow Process(%d) failed:%v\n", idx, err)
				return
			} else {
				fmt.Printf("DataFlow Process(%d)=%v\n", idx, ret)
			}
		}(iLoop)
	}
	wg.Wait()
	df.Destroy()
}
