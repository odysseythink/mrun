package metrics

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func BenchmarkGuage(b *testing.B) {
	g := NewGauge[int64]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(int64(i))
	}
}

// exercise race detector
func TestGaugeConcurrency(t *testing.T) {
	rand.Seed(time.Now().Unix())
	g := NewGauge[int64]()
	wg := &sync.WaitGroup{}
	reps := 100
	for i := 0; i < reps; i++ {
		wg.Add(1)
		go func(g Gauge[int64], wg *sync.WaitGroup) {
			g.Update(rand.Int63())
			wg.Done()
		}(g, wg)
	}
	wg.Wait()
}

func TestGauge(t *testing.T) {
	g := NewGauge[int64]()
	g.Update(int64(47))
	if v := g.Value(); 47 != v {
		t.Errorf("g.Value(): 47 != %v\n", v)
	}
}

func TestGaugeSnapshot(t *testing.T) {
	g := NewGauge[int64]()
	g.Update(int64(47))
	snapshot := g.Snapshot()
	g.Update(int64(0))
	if v := snapshot.Value(); 47 != v {
		t.Errorf("g.Value(): 47 != %v\n", v)
	}
}

func TestGetOrRegisterGauge(t *testing.T) {
	r := NewRegistry()
	NewRegisteredGauge[int64]("foo", r).Update(47)
	if g := GetOrRegisterGauge[int64]("foo", r); 47 != g.Value() {
		t.Fatal(g)
	}
}

func ExampleGetOrRegisterGauge() {
	m := "server.bytes_sent"
	g := GetOrRegisterGauge[int64](m, nil)
	g.Update(47)
	fmt.Println(g.Value()) // Output: 47
}
