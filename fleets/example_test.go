package fleets

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	sum int32
	wg  sync.WaitGroup
)

func incSum(i any) {
	incSumInt(i.(int32))
}

func incSumInt(i int32) {
	atomic.AddInt32(&sum, i)
	wg.Done()
}

func ExamplePool() {
	Reboot() // ensure the default pool is available

	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)
	// Use the default pool.
	for i := 0; i < runTimes; i++ {
		j := i
		_ = Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	fmt.Printf("The result is %d\n", sum)

	atomic.StoreInt32(&sum, 0)
	wg.Add(runTimes)
	// Use the new pool.
	pool, _ := NewPool(10)
	defer pool.Release()
	for i := 0; i < runTimes; i++ {
		j := i
		_ = pool.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	fmt.Printf("The result is %d\n", sum)

	// Output:
	// The result is 499500
	// The result is 499500
}

func ExamplePoolWithFunc() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	pool, _ := NewPoolWithFunc(10, incSum)
	defer pool.Release()

	for i := 0; i < runTimes; i++ {
		_ = pool.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExamplePoolWithFuncGeneric() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	pool, _ := NewPoolWithFuncGeneric(10, incSumInt)
	defer pool.Release()

	for i := 0; i < runTimes; i++ {
		_ = pool.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExampleMultiPool() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	mp, _ := NewMultiPool(10, runTimes/10, RoundRobin)
	defer mp.ReleaseTimeout(time.Second) // nolint:errcheck

	for i := 0; i < runTimes; i++ {
		j := i
		_ = mp.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExampleMultiPoolWithFunc() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	mp, _ := NewMultiPoolWithFunc(10, runTimes/10, incSum, RoundRobin)
	defer mp.ReleaseTimeout(time.Second) // nolint:errcheck

	for i := 0; i < runTimes; i++ {
		_ = mp.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExampleMultiPoolWithFuncGeneric() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	mp, _ := NewMultiPoolWithFuncGeneric(10, runTimes/10, incSumInt, RoundRobin)
	defer mp.ReleaseTimeout(time.Second) // nolint:errcheck

	for i := 0; i < runTimes; i++ {
		_ = mp.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}
