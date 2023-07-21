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
	pdf := mrun.NewDataFlow("parallel")
	sdf := mrun.NewDataFlow("sequence")
	err := sdf.Register(&Processor1{}, []mrun.DataFlowOption{mrun.NewDataFlowOrderOption(0)}, nil)
	if err != nil {
		fmt.Println("sequence DataFlow Register Processor1 failed:", err)
		return
	}
	err = pdf.Register(&Processor1{}, nil, nil)
	if err != nil {
		fmt.Println("parallel DataFlow Register Processor1 failed:", err)
		return
	}
	err = sdf.Register(&Processor2{}, []mrun.DataFlowOption{mrun.NewDataFlowOrderOption(1)}, nil)
	if err != nil {
		fmt.Println("sequence DataFlow Register Processor2 failed:", err)
		return
	}
	err = pdf.Register(&Processor2{}, nil, nil)
	if err != nil {
		fmt.Println("parallel DataFlow Register Processor1 failed:", err)
		return
	}
	err = sdf.Init()
	if err != nil {
		fmt.Println("sequence DataFlow init failed:", err)
		return
	}
	err = pdf.Init()
	if err != nil {
		fmt.Println("parallel DataFlow init failed:", err)
		return
	}
	var wg sync.WaitGroup
	for iLoop := 0; iLoop < 10; iLoop++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ret, err := sdf.Process(idx)
			if err != nil {
				fmt.Printf("sequence DataFlow Process(%d) failed:%v\n", idx, err)
				return
			} else {
				fmt.Printf("sequence DataFlow Process(%d)=%v\n", idx, ret)
			}
		}(iLoop)

		ret, err := pdf.Process(iLoop)
		if err != nil {
			fmt.Printf("parallel DataFlow Process(%d) failed:%v\n", iLoop, err)
			return
		} else {
			fmt.Printf("parallel DataFlow Process(%d)=%v\n", iLoop, ret)
		}
	}
	wg.Wait()
	sdf.Destroy()
	pdf.Destroy()
}
