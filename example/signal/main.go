package main

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"mlib.com/mrun"
)

type MainSystem struct {
	msgSig *mrun.Signal
	wg     *atomic.Int32
}

func (s *MainSystem) Init(args ...any) error {
	return nil
}
func (s *MainSystem) Destroy() {

}
func (s *MainSystem) RunOnce(ctx context.Context) error {
	timer := time.NewTimer(10 * time.Millisecond)
	select {
	case <-ctx.Done():
		return nil
	case <-timer.C:
		// log.Println("------s.msgSig=", s.msgSig)
		if s.wg.Load() == 0 {
			for iLoop := range 10000 {
				s.wg.Add(1)
				go s.msgSig.Emit("message_was_created", strconv.Itoa(iLoop))
			}
		}

		return nil
	}
	return nil
}
func (s *MainSystem) UserData() any {
	return nil
}

func (s *MainSystem) F1(val string) string {
	fmt.Println("------hello world=", val)
	if s.wg.Load() > 0 {
		s.wg.Add(-1)
	}
	return ""
}

func main() {
	// // 定义一个匿名字段的基类型
	// baseType := reflect.TypeOf(struct {
	// 	Name string
	// }{})

	// baseType1 := reflect.TypeOf(func(string) string { return "hello" })
	// type Cbfunc func(string) string
	// baseType2 := reflect.TypeOf([]Cbfunc{func(string) string { return "hello" }})

	// // 创建一个包含匿名字段的结构体类型
	// myType := reflect.StructOf([]reflect.StructField{
	// 	{
	// 		Name:      "Base", // 匿名字段的类型名作为字段名
	// 		Type:      baseType,
	// 		Anonymous: true, // 设置为匿名字段
	// 	},
	// 	{
	// 		Name: "Age",
	// 		Type: reflect.TypeOf(int(0)),
	// 	},
	// 	{
	// 		Name:      "Cb", // 匿名字段的类型名作为字段名
	// 		Type:      baseType1,
	// 		Anonymous: true, // 设置为匿名字段
	// 	},
	// 	{
	// 		Name:      "Cbs", // 匿名字段的类型名作为字段名
	// 		Type:      baseType2,
	// 		Anonymous: true, // 设置为匿名字段
	// 	},
	// })

	// // 使用反射创建结构体实例
	// instance := reflect.New(myType).Elem()

	// // 设置字段值
	// instance.Field(0).FieldByName("Name").SetString("John") // 设置匿名字段的Name属性
	// instance.Field(1).SetInt(30)                            // 设置Age字段
	// instance.Field(2).Set(reflect.ValueOf(F1))
	// cbvalue := instance.FieldByName("Cb")
	// if !cbvalue.IsNil() {
	// 	rsp := instance.Field(2).Call([]reflect.Value{reflect.ValueOf("1111")})
	// 	fmt.Printf("response: %+v\n", rsp[0])
	// }
	// {
	// 	realcbsvalue := instance.FieldByName("Cbs").Interface().([]Cbfunc)
	// 	realcbsvalue = append(realcbsvalue, F1)
	// 	instance.FieldByName("Cbs").Set(reflect.ValueOf(realcbsvalue))
	// 	cbsvalue := instance.FieldByName("Cbs")
	// 	fmt.Printf("len=: %d\n", cbsvalue.Len())
	// 	if !cbsvalue.IsNil() {
	// 		if cbsvalue.Len() > 0 {
	// 			rsp := cbsvalue.Index(0).Call([]reflect.Value{reflect.ValueOf("1111")})
	// 			fmt.Printf("-----response: %+v\n", rsp[0])
	// 		}
	// 	}
	// }

	// // 输出结构体实例的值
	// fmt.Printf("value: %+v\n", instance.Interface())

	// st := reflect.TypeOf(Signal{})
	// num := st.NumField()
	// newstfields := []reflect.StructField{}
	// for i := 0; i < num; i++ {
	// 	f := st.Field(i)
	// 	newstfields = append(newstfields, f)
	// }
	// myType1 := reflect.StructOf(newstfields)
	// instance1 := reflect.New(myType1).Elem()
	// fmt.Printf("value1: %+v\n", instance1.Interface())

	// fmt.Println("------type(F1) == type(F2)", reflect.TypeOf(F1) == reflect.TypeOf(F2))
	msys := &MainSystem{
		wg: &atomic.Int32{},
	}
	msys.wg.Store(0)

	msgSig, err := mrun.NewSignal("message_was_created", func(string) string { return "" }, mrun.NewSignalConcurrencyOption(100))
	if err != nil {
		fmt.Println("new signal failed:", err)
		return
	}
	msys.msgSig = msgSig
	err = mrun.Connect(msgSig, msys.F1)
	if err != nil {
		fmt.Println("connect signal failed:", err)
		return
	}
	msys.msgSig.EmitDirect("message_was_created", "hello")
	msys.msgSig.Emit("message_was_created", "111")
	mrun.Run(msys, syscall.SIGINT, syscall.SIGTERM)
}
